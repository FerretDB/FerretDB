function initCodapi() {
    setTimeout(() => {
        document.querySelectorAll("codapi-snippet").forEach((el) => {
            const snippet = document.createElement("codapi-snippet");
            setAttribute(snippet, el, "sandbox");
            setAttribute(snippet, el, "editor");
            setAttribute(snippet, el, "template");
            el.replaceWith(snippet);
        });
    }, 500);
}

function setAttribute(dst, src, attrName) {
    if (!src.hasAttribute(attrName)) {
        return;
    }
    dst.setAttribute(attrName, src.getAttribute(attrName));
}

addEventListener("load", initCodapi);