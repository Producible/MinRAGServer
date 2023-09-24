// static/script.js
document.addEventListener('DOMContentLoaded', (event) => {
    const treeView = document.querySelector('.tree-view');
    treeView.addEventListener('click', (event) => {
        const item = event.target.closest('li');
        if (item && event.target.tagName === 'SPAN') {
            item.classList.toggle('expanded');
        }
    });
});
