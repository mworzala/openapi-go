# openapi-go
An opinionated openapi server code generator for go

todo format me + add more

Below are the additional spec properties added
* `info.x-base-path` - path for the server excl. version (default: `/{specName}`, example: `/abc` -> `/v1/abc`)
* `paths.{path}.{method}.responses.{code}.x-type` - set to `empty` to use that code as the empty response. 
  That code may not have a response body.

Object properties are treated as required unless the individual field has `required: false` or 
the object root has `required: false`.
