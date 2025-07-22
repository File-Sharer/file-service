# file-service

## API Docs
`/api` - base route

**Headers**:
- **`Authorization`**: Bearer `<ACCESS_TOKEN>`

**Designations**:
- **`[AUTH]`** - ***requires** auth*
- **`[X_INTERNAL_TOKEN]`** - ***requires** internal token*

**`[X_INTERNAL_TOKEN]`** `/users-spaces`:
- **PATCH** -> `/level` - *update user space level*

**`[AUTH]`** `/files`:
- **POST** -> `/` - *create a file*
- **GET** -> `/:<file_id>` - *get file by ID*
- **GET** -> `/` - *get your own files*
- **GET** -> `/:<file_id>/dl` - *download file*
- **PUT** -> `/:<file_id>/:<user_id>` - *add permission to file*
- **DELETE** -> `/:<file_id>` - *delete file*
- **DELETE** -> `/:<file_id>/:<user_id>` - *delete permission*
- **GET** -> `/:<file_id>/permissions` - *get permissions to your file*
- *PATCH* -> `/:<file_id>/togglepub` - *toggle file visibility*
