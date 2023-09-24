// static/clipboard.js

function showToast(message) {
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => {
        document.body.removeChild(toast);
    }, 2000);  // Toast duration
}

function copyToClipboard(text) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
    showToast('URL copied to clipboard');  // Show toast instead of alert
}

document.addEventListener('click', (event) => {
    const copyButton = event.target.closest('.copy-button');
    const copyInfoButton = event.target.closest('.copy-button-info');
    if (copyButton) {
        const url = copyButton.getAttribute('data-url');
        copyToClipboard(url);
    } else if (copyInfoButton) {
        const info = copyInfoButton.getAttribute('data-info');
        copyToClipboard(info);
    }
});
