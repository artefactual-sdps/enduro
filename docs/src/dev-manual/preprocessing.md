# Preprocessing child workflow

The processing workflow can be extended with the execution of a preprocessing
child workflow.

## Configuration

### `.tilt.env`

Check the [Tilt environment configuration].

### `enduro.toml`

```toml
# Optional preprocessing child workflow configuration.
[preprocessing]
# enabled triggers the execution of the child workflow, when set to false all other
# options are ignored.
enabled = true
# extract determines if the package extraction happens on the child workflow.
extract = false
# sharedPath is the full path to the directory used to share the package between workflows,
# required when enabled is set to true.
sharedPath = "/home/enduro/preprocessing"

# Temporal configuration to trigger the preprocessing child workflow, all fields are
# required when enabled is set to true.
[preprocessing.temporal]
namespace = "default"
taskQueue = "preprocessing"
workflowName = "preprocessing"
```

[tilt environment configuration]: devel.md#preprocessing_path
