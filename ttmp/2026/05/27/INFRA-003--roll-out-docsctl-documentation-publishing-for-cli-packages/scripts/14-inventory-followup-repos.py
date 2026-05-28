#!/usr/bin/env python3
from pathlib import Path
import json, re, subprocess
base=Path('/home/manuel/code/wesen/go-go-golems')
skip={'.git','.github','ttmp'}
rows=[]
for d in sorted([p for p in base.iterdir() if p.is_dir() and p.name not in skip], key=lambda p:p.name):
    gomod=d/'go.mod'
    if not gomod.exists():
        continue
    txt=gomod.read_text(errors='ignore')
    module=''
    m=re.search(r'^module\s+(\S+)', txt, re.M)
    if m: module=m.group(1)
    make=(d/'Makefile').read_text(errors='ignore') if (d/'Makefile').exists() else ''
    workflows='\n'.join(p.read_text(errors='ignore') for p in (d/'.github'/'workflows').glob('*') if p.is_file()) if (d/'.github'/'workflows').exists() else ''
    gofiles=list(d.rglob('*.go'))
    # prune common huge/historical dirs for feature markers
    gofiles=[p for p in gofiles if '/.git/' not in str(p) and '/ttmp/' not in str(p) and '/vendor/' not in str(p)]
    def grep_files(pattern):
        r=re.compile(pattern)
        for p in gofiles:
            try:
                if r.search(p.read_text(errors='ignore')): return True
            except Exception: pass
        return False
    has_cmd=(d/'cmd').exists() or (d/'main.go').exists()
    docs_help=(d/'docs'/'help').exists() or (d/'help').exists()
    row={
      'repo':d.name,'path':str(d),'module':module,'has_cmd':has_cmd,
      'uses_glazed':'github.com/go-go-golems/glazed' in txt,
      'uses_goja':'github.com/go-go-golems/go-go-goja' in txt,
      'uses_logcopter':'github.com/go-go-golems/logcopter' in txt,
      'has_logcopter_generate':(d/'logcopter_generate.go').exists(),
      'logcopter_go_count':sum(1 for p in gofiles if p.name=='logcopter.go'),
      'has_logcopter_check':'logcopter-check' in make,
      'has_bump_go_go_golems':'bump-go-go-golems' in make,
      'has_glazed_lint':'glazed-lint' in make,
      'has_docsctl_ci':'publish-docsctl' in workflows or 'docsctl publish' in workflows,
      'has_release_workflow': any((d/'.github'/'workflows'/name).exists() for name in ['release.yaml','release.yml']),
      'has_help_export': grep_files(r'help\s+export|SetupCobraRootCommand|help_cmd\.Setup|LoadSectionsFromFS|help export'),
      'docs_help_dir': docs_help,
      'has_xgoja_provider_dir': any(p.is_dir() and (str(p).endswith('/pkg/xgoja/provider') or re.search(r'/pkg/js/modules/[^/]+/provider$', str(p))) for p in d.rglob('*') if p.is_dir()),
      'registers_goja_module': grep_files(r'modules\.Register|NativeModule|runtimebridge|NewRegistrar'),
    }
    # heuristic needs
    row['needs_logcopter_addition']= not (row['has_logcopter_generate'] and row['has_logcopter_check'] and row['logcopter_go_count']>0)
    row['needs_glazed_linting_added']= row['uses_glazed'] and not row['has_glazed_lint']
    row['docsctl_candidate']= row['has_cmd'] and row['uses_glazed'] and (row['has_help_export'] or row['docs_help_dir'])
    row['needs_docsctl_cicd_push']= row['docsctl_candidate'] and not row['has_docsctl_ci']
    row['xgoja_candidate']= row['uses_goja'] and row['registers_goja_module']
    row['needs_xgoja_bindings']= row['uses_goja'] and row['registers_goja_module'] and not row['has_xgoja_provider_dir']
    rows.append(row)
print(json.dumps(rows, indent=2))
