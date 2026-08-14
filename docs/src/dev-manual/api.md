<!-- markdownlint-disable MD033 MD041 -->
<!--
  MD033: no-inline-html
  MD041: first-line-heading/first-line-h1
-->
The Enduro API contract is available in both
[OpenAPI 3.2](api/openapi3.2.json) and
[OpenAPI 3.0](api/openapi3.json) formats. OpenAPI 3.2 is the primary contract;
the 3.0 document is retained for tools that do not yet support 3.2.

Both documents are generated from the Goa design by running `make gen-goa`.

<swagger-ui src="api/openapi3.2.json"/>
