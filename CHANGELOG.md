# Changelog

## [1.0.3](https://github.com/nguyenvanduocit/confluence-mcp/compare/v1.0.2...v1.0.3) (2026-03-13)


### Bug Fixes

* add explicit archive IDs for homebrew brew formula ([c210e77](https://github.com/nguyenvanduocit/confluence-mcp/commit/c210e7782aae17300d511874649b39dacbd4d4df))

## [1.0.2](https://github.com/nguyenvanduocit/confluence-mcp/compare/v1.0.1...v1.0.2) (2026-03-13)


### Bug Fixes

* checkout tagged commit in goreleaser to fix tag mismatch ([50f6561](https://github.com/nguyenvanduocit/confluence-mcp/commit/50f6561e5b6cf9c82d2ee6c1cf101b5e2f420712))

## [1.0.1](https://github.com/nguyenvanduocit/confluence-mcp/compare/v1.0.0...v1.0.1) (2026-03-13)


### Bug Fixes

* use merge instead of squash to preserve commit SHAs for release-please ([f26b66f](https://github.com/nguyenvanduocit/confluence-mcp/commit/f26b66f2e5f24e5f13d72afe1b01b90c2b26cd9b))

## 1.0.0 (2026-03-13)


### Features

* add CLI package and update README with CLI usage ([c9f5a8c](https://github.com/nguyenvanduocit/confluence-mcp/commit/c9f5a8c0d021939c803f8bd71e85294465cad1e3))
* add Gitleaks workflow for secret scanning ([c1c0fb4](https://github.com/nguyenvanduocit/confluence-mcp/commit/c1c0fb43b4c6b48ff6cf8a964ab956164a9ff7a0))
* add homebrew installation instructions ([e8d13b7](https://github.com/nguyenvanduocit/confluence-mcp/commit/e8d13b73f3e09aa0f1f33f887b380ace4b0e848a))
* **api:** add Confluence page comments functionality ([4aafff4](https://github.com/nguyenvanduocit/confluence-mcp/commit/4aafff4e5fbc4296189b7ac48647ec2a983cbf8a))
* **confluence:** split monolithic tool into modular Confluence tools\n\n- Remove tools/confluence.go\n- Add tools/confluence_search.go, tools/confluence_page.go, tools/confluence_create.go, tools/confluence_update.go\n- Update main.go to register new modular tools ([53c4031](https://github.com/nguyenvanduocit/confluence-mcp/commit/53c4031472aac9a9f52457481feb97a3dd550312))
* **docker:** add Dockerfile and update README for Docker usage\n\n- Add Dockerfile for containerized builds\n- Update README with Docker instructions and environment variable usage\n- Improve main.go to check required envs for Docker/production ([56472d9](https://github.com/nguyenvanduocit/confluence-mcp/commit/56472d9e207e82792bebc53262e4999c43f20e19))
* **docker:** add GitHub Container Registry support ([ed034e0](https://github.com/nguyenvanduocit/confluence-mcp/commit/ed034e071c737889ce26911d8e76262981690e1e))
* init ([0db1a57](https://github.com/nguyenvanduocit/confluence-mcp/commit/0db1a575842a6426d18b1dc8d85de26ac7f9c187))
* **readme:** enhance README with detailed project description and add thumbnail image ([2a79386](https://github.com/nguyenvanduocit/confluence-mcp/commit/2a793868f3d11c528aac0d4fee27bc84175e7e8f))
* remove sse, and support streamableHttpServer ([85f000d](https://github.com/nguyenvanduocit/confluence-mcp/commit/85f000d50969bfcb9acc805c92eec59f68ee5caf))


### Bug Fixes

* add --repo flag to gh pr merge to fix auto-merge ([1b270f6](https://github.com/nguyenvanduocit/confluence-mcp/commit/1b270f63f68c075eccc1d256aa934bb8fa999fce))
* add concurrency control to prevent release race conditions ([da3c263](https://github.com/nguyenvanduocit/confluence-mcp/commit/da3c2636c4d636b5df2d06d7a91a65977649a396))
* **core:** improve error handling for Confluence client initialization ([c445070](https://github.com/nguyenvanduocit/confluence-mcp/commit/c445070f899a267ad9e758b46c75aa6fe00474ea))

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
