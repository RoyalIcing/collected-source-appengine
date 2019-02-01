
- Content management system
- Write Markdown
- Upload CSV
- 

## People

- Developers who need a CMS for their website
- Front-end developers who need a backend for their single-page app
- Members of teams who need to communicate with the other members
- Teams who need to communicate with other teams

## Repos

### Content

Content is stored by their content address.

### Taxonomy

Has many tags, which can be used to search by.

Has many referencing UUIDs, which can be used to show replies in order.

### Timestamps

Created at

### Item model

- Item
  - groupUUID
  - uuid
  - Content
    - sha256
  - Timestamps
    - createdAt
  - Taxonomoy
    - tags
    - referencingUUID
