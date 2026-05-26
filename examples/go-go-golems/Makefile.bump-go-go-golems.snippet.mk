# Generic go-go-golems dependency bump target.
#
# Copy into a Go repository Makefile when the repository directly depends on
# go-go-golems modules. It scans go.mod instead of maintaining a stale manual
# dependency list, then bumps every direct github.com/go-go-golems/... module to
# @latest and tidies the module.

.PHONY: bump-go-go-golems
bump-go-go-golems:
	@deps="$$(awk '/^require[[:space:]]+github\.com\/go-go-golems\// { print $$2 } /^[[:space:]]*github\.com\/go-go-golems\// { print $$1 }' go.mod | sort -u)"; \
	if [ -z "$$deps" ]; then \
		echo "No github.com/go-go-golems dependencies in go.mod"; \
	else \
		echo "Bumping go-go-golems dependencies:"; \
		echo "$$deps"; \
		for dep in $$deps; do go get "$${dep}@latest"; done; \
	fi
	go mod tidy
