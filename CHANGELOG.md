# Changelog

## [1.4.0](https://github.com/nguyenvanduocit/confluence-mcp/compare/v1.3.0...v1.4.0) (2025-04-18)


### Features

* **api:** add Confluence page comments functionality ([36f7e83](https://github.com/nguyenvanduocit/confluence-mcp/commit/36f7e83cdc959983d39e53350f3af0669416fa29))
* **docker:** add GitHub Container Registry support ([9f3a094](https://github.com/nguyenvanduocit/confluence-mcp/commit/9f3a094bcc714bcb1b3182d9aeb26ded6c9c4344))

## [1.3.0](https://github.com/nguyenvanduocit/confluence-mcp/compare/v1.2.0...v1.3.0) (2025-04-17)


### Features

* add Gitleaks workflow for secret scanning ([c1c0fb4](https://github.com/nguyenvanduocit/confluence-mcp/commit/c1c0fb43b4c6b48ff6cf8a964ab956164a9ff7a0))
* **confluence:** split monolithic tool into modular Confluence tools\n\n- Remove tools/confluence.go\n- Add tools/confluence_search.go, tools/confluence_page.go, tools/confluence_create.go, tools/confluence_update.go\n- Update main.go to register new modular tools ([1010f91](https://github.com/nguyenvanduocit/confluence-mcp/commit/1010f910949983d5d981917fbccb035fe966f4ed))
* **docker:** add Dockerfile and update README for Docker usage\n\n- Add Dockerfile for containerized builds\n- Update README with Docker instructions and environment variable usage\n- Improve main.go to check required envs for Docker/production ([bceff46](https://github.com/nguyenvanduocit/confluence-mcp/commit/bceff46e1c4b7a99e5f61ce15e299b9f6f3984b6))
* init ([0db1a57](https://github.com/nguyenvanduocit/confluence-mcp/commit/0db1a575842a6426d18b1dc8d85de26ac7f9c187))

## [1.2.0](https://github.com/nguyenvanduocit/confluence-mcp/compare/v1.1.0...v1.2.0) (2025-04-17)


### Features

* **confluence:** split monolithic tool into modular Confluence tools\n\n- Remove tools/confluence.go\n- Add tools/confluence_search.go, tools/confluence_page.go, tools/confluence_create.go, tools/confluence_update.go\n- Update main.go to register new modular tools ([1010f91](https://github.com/nguyenvanduocit/confluence-mcp/commit/1010f910949983d5d981917fbccb035fe966f4ed))

## [1.1.0](https://github.com/nguyenvanduocit/confluence-mcp/compare/v1.0.0...v1.1.0) (2025-04-17)


### Features

* **docker:** add Dockerfile and update README for Docker usage\n\n- Add Dockerfile for containerized builds\n- Update README with Docker instructions and environment variable usage\n- Improve main.go to check required envs for Docker/production ([bceff46](https://github.com/nguyenvanduocit/confluence-mcp/commit/bceff46e1c4b7a99e5f61ce15e299b9f6f3984b6))

## 1.0.0 (2025-03-25)


### Features

* add Gitleaks workflow for secret scanning ([c1c0fb4](https://github.com/nguyenvanduocit/confluence-mcp/commit/c1c0fb43b4c6b48ff6cf8a964ab956164a9ff7a0))
* init ([0db1a57](https://github.com/nguyenvanduocit/confluence-mcp/commit/0db1a575842a6426d18b1dc8d85de26ac7f9c187))
