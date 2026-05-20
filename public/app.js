const fileList = document.querySelector("#fileList");
const message = document.querySelector("#message");
const uploadForm = document.querySelector("#uploadForm");
const fileInput = document.querySelector("#fileInput");
const refreshButton = document.querySelector("#refreshButton");

const formatBytes = (bytes) => {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 1)} ${units[index]}`;
};

const setMessage = (text, isError = false) => {
  message.textContent = text;
  message.style.color = isError ? "#b93939" : "#4d626b";
};

async function loadFiles() {
  const response = await fetch("api/files");
  if (!response.ok) {
    setMessage("โหลดรายการไฟล์ไม่สำเร็จ", true);
    return;
  }

  const files = await response.json();
  fileList.innerHTML = "";

  if (files.length === 0) {
    fileList.innerHTML = '<div class="empty">ยังไม่มีไฟล์</div>';
    return;
  }

  for (const file of files) {
    const row = document.createElement("article");
    row.className = "file-row";
    row.innerHTML = `
      <div>
        <div class="file-name"></div>
        <div class="file-meta">${formatBytes(file.size)} · ${new Date(file.updated_at).toLocaleString()}</div>
      </div>
      <a class="link-button" href="${file.url}">Download</a>
      <button class="danger" type="button">Delete</button>
    `;

    row.querySelector(".file-name").textContent = file.name;
    row.querySelector("button").addEventListener("click", async () => {
      if (!confirm(`Delete ${file.name}?`)) return;
      const response = await fetch(`api/files/${encodeURIComponent(file.name)}`, { method: "DELETE" });
      if (!response.ok) {
        setMessage("ลบไฟล์ไม่สำเร็จ", true);
        return;
      }
      setMessage("ลบไฟล์แล้ว");
      await loadFiles();
    });

    fileList.appendChild(row);
  }
}

uploadForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = new FormData(uploadForm);

  setMessage("กำลังอัปโหลด...");
  const response = await fetch("api/files", {
    method: "POST",
    body: data,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: "upload failed" }));
    setMessage(error.message ?? "อัปโหลดไม่สำเร็จ", true);
    return;
  }

  fileInput.value = "";
  setMessage("อัปโหลดสำเร็จ");
  await loadFiles();
});

refreshButton.addEventListener("click", loadFiles);
loadFiles();
