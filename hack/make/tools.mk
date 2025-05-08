# List of tools.
TOOLS = \
    atlas \
    ent \
    go-enum \
    goa \
    golangci-lint \
    gomajor \
    gotestsum \
    migrate \
    mockgen \
    shfmt \
    tparse \
    workflowcheck

# Pattern rule to install each tool.
tool-%:
	@go tool bine get $* 1> /dev/null
