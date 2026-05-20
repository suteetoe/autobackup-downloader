# File Download Server with Go Echo

โปรเจกต์ตัวอย่างสำหรับอัปโหลด ลิสต์ ดาวน์โหลด และลบไฟล์ด้วย Golang + Echo

## Run

```bash
go mod tidy
go run .
```

เปิดเว็บที่:

```text
http://localhost:8080
```

ถ้าพอร์ต `8080` ถูกใช้งานอยู่ สามารถเปลี่ยนพอร์ตได้:

```bash
PORT=8090 go run .
```

## Config

ตั้งค่าผ่าน environment variables:

| Name | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | port สำหรับ HTTP server |
| `FILE_STORAGE_DIR` | `files` | folder สำหรับเก็บไฟล์ upload/download |

ตัวอย่างรันโดยชี้ folder เก็บไฟล์เอง:

```bash
FILE_STORAGE_DIR=/data/downloads PORT=8090 go run .
```

ตัวอย่าง mount volume เมื่อรันผ่าน Docker:

```bash
docker run --rm \
  -p 8090:8080 \
  -e FILE_STORAGE_DIR=/app/storage \
  -v "$PWD/storage:/app/storage" \
  your-image-name
```

## Docker

Build image:

```bash
docker build -t file-download-server .
```

Run container:

```bash
docker run --rm \
  -p 8090:8080 \
  -v "$PWD/storage:/app/storage" \
  file-download-server
```

เปิดเว็บที่:

```text
http://localhost:8090
```

## GitHub Container Registry

GitHub Actions จะ build Docker image และ push ไปที่ GitHub Container Registry เมื่อ push เข้า `main`/`master`, push tag แบบ `v*.*.*`, หรือกดรันเองผ่าน `workflow_dispatch`

Image name:

```text
ghcr.io/<owner>/<repository>
```

ตัวอย่าง pull:

```bash
docker pull ghcr.io/<owner>/<repository>:latest
```

Pull request จะ build เพื่อตรวจสอบอย่างเดียว แต่จะไม่ push image เข้า registry

## API

```http
GET /api/files
POST /api/files
GET /api/files/:name/download
DELETE /api/files/:name
```

ตัวอย่างอัปโหลดผ่าน `curl`:

```bash
curl -F "file=@example.pdf" http://localhost:8080/api/files
```

ไฟล์ที่อัปโหลดจะถูกเก็บในโฟลเดอร์ `files/`
