package release

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type verifyDocsSettings struct {
	Package     string
	Version     string
	BaseURL     string
	MinSections int
	Timeout     time.Duration
	Output      string
}

type verifyDocsResult struct {
	OK           bool   `json:"ok" yaml:"ok"`
	URL          string `json:"url" yaml:"url"`
	Package      string `json:"package" yaml:"package"`
	Version      string `json:"version" yaml:"version"`
	StatusCode   int    `json:"status_code" yaml:"status_code"`
	Title        string `json:"title,omitempty" yaml:"title,omitempty"`
	SectionCount int    `json:"section_count" yaml:"section_count"`
	HasPackage   bool   `json:"has_package" yaml:"has_package"`
	HasVersion   bool   `json:"has_version" yaml:"has_version"`
	HasPreload   bool   `json:"has_preloaded_state" yaml:"has_preloaded_state"`
	Error        string `json:"error,omitempty" yaml:"error,omitempty"`
}

func newVerifyDocsCommand() *cobra.Command {
	s := &verifyDocsSettings{BaseURL: "https://docs.yolo.scapegoat.dev", MinSections: 1, Timeout: 30 * time.Second, Output: "table"}
	cmd := &cobra.Command{
		Use:   "verify-docs",
		Short: "Verify that a docsctl-published package/version is visible in the docs browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			res := verifyDocs(cmd.Context(), s)
			if err := writeVerifyDocsResult(res, s.Output); err != nil {
				return err
			}
			if !res.OK {
				return fmt.Errorf("docs verification failed for %s@%s", s.Package, s.Version)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&s.Package, "package", "", "Package name to verify")
	cmd.Flags().StringVar(&s.Version, "version", "", "Package version to verify")
	cmd.Flags().StringVar(&s.BaseURL, "base-url", s.BaseURL, "Docs browser base URL")
	cmd.Flags().IntVar(&s.MinSections, "min-sections", s.MinSections, "Minimum expected section count")
	cmd.Flags().DurationVar(&s.Timeout, "timeout", s.Timeout, "HTTP timeout")
	cmd.Flags().StringVar(&s.Output, "output", s.Output, "Output format: table, json, yaml")
	_ = cmd.MarkFlagRequired("package")
	_ = cmd.MarkFlagRequired("version")
	return cmd
}

func verifyDocs(ctx context.Context, s *verifyDocsSettings) verifyDocsResult {
	base := strings.TrimRight(s.BaseURL, "/")
	url := fmt.Sprintf("%s/%s/%s", base, s.Package, s.Version)
	res := verifyDocsResult{URL: url, Package: s.Package, Version: s.Version}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	httpRes, err := http.DefaultClient.Do(req)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer httpRes.Body.Close()
	res.StatusCode = httpRes.StatusCode
	bodyBytes, err := io.ReadAll(io.LimitReader(httpRes.Body, 4<<20))
	if err != nil {
		res.Error = err.Error()
		return res
	}
	body := string(bodyBytes)
	res.Title = firstSubmatch(body, regexp.MustCompile(`(?is)<title>(.*?)</title>`))
	res.HasPreload = strings.Contains(body, "window.__PRELOADED_STATE__")
	res.HasPackage = strings.Contains(body, `"name":"`+s.Package+`"`) || strings.Contains(body, `"packageName":"`+s.Package+`"`)
	res.HasVersion = strings.Contains(body, `"`+s.Version+`"`) || strings.Contains(body, s.Package+" "+s.Version)
	if section := packageSectionCount(body, s.Package); section != "" {
		res.SectionCount, _ = strconv.Atoi(section)
	} else if total := firstSubmatch(body, regexp.MustCompile(`"total":(\d+)`)); total != "" {
		res.SectionCount, _ = strconv.Atoi(total)
	}
	res.OK = res.StatusCode >= 200 && res.StatusCode < 300 && res.HasPackage && res.HasVersion && res.SectionCount >= s.MinSections
	return res
}

func firstSubmatch(s string, re *regexp.Regexp) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func packageSectionCount(body, packageName string) string {
	pattern := fmt.Sprintf(`\{"name":%q.*?"sectionCount":(\d+)`, regexp.QuoteMeta(packageName))
	return firstSubmatch(body, regexp.MustCompile(pattern))
}

func writeVerifyDocsResult(res verifyDocsResult, output string) error {
	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	case "yaml":
		b, err := yaml.Marshal(res)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		return err
	default:
		fmt.Printf("ok\tpackage\tversion\tsections\tstatus\turl\n")
		fmt.Printf("%t\t%s\t%s\t%d\t%d\t%s\n", res.OK, res.Package, res.Version, res.SectionCount, res.StatusCode, res.URL)
		if res.Error != "" {
			fmt.Fprintf(os.Stderr, "error: %s\n", res.Error)
		}
		return nil
	}
}
