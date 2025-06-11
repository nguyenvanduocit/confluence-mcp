# Confluence MCP – Implementation Checklist

This file tracks the design & implementation work for the next wave of Confluence MCP tools.  Update the check-boxes as work progresses and feel free to add notes, PR links, or owners.

## Legend
- [ ] TODO – not started
- [/] IN PROGRESS – currently being worked on
- [x] DONE – feature merged & released

---

## Navigation & Discovery
- [x] **ListSpacesTool** – list all Confluence spaces available to the user
- [ ] **GetSpaceHomepageTool** – fetch the homepage of a given space
- [ ] **ListChildPagesTool** – return the child pages of a parent page (page tree)
- [ ] **GetPageVersionsTool** – list all versions of a page
- [ ] **RestorePageVersionTool** – roll back a page to a selected version
- [ ] **SearchAttachmentsTool** – search only attachments within Confluence

## Collaboration & Content Creation
- [ ] **AddCommentTool** – add a new page or inline comment
- [ ] **UpdateCommentTool** – edit an existing comment
- [ ] **DeleteCommentTool** – remove a comment
- [ ] **UploadAttachmentTool** – upload an attachment to a page
- [ ] **DownloadAttachmentTool** – download/stream an attachment
- [ ] **AddLabelTool** – add labels to a page or attachment
- [ ] **CreateDraftPageTool** – create a draft page without publishing

## Governance & House-Keeping
- [ ] **MovePageTool** – move/re-parent or reorder a page
- [ ] **DeletePageTool** – remove a page (soft delete)
- [ ] **ArchivePageTool** – archive pages to a designated archive space and label them
- [ ] **WatchPageTool** – subscribe the user to change notifications on a page
- [ ] **GetSpacePermissionsTool** – retrieve permission details for a space
- [ ] **BulkLabelTool** – add or remove a label across multiple pages in bulk

---