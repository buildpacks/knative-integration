### Knative Cloud Native Buildpacks Integration


### Run the test locally
Need to update the builder.toml file to add the following lines at the end of the file:
```toml
run-image = "cnbs/run@$RUN_IMAGE_DIGEST"
build-image = "cnbs/build@$BUILD_IMAGE_DIGEST"

[lifecycle]
uri = "file://$LIFECYCLE_BINARIES"
```
where:

`$RUN_IMAGE_DIGEST` is the digest of the latest cnbs/run

`$BUILD_IMAGE_DIGEST` is the digest of the latest cnbs/build

`$LIFECYCLE_BINARIES` path to the lifecycle binaries