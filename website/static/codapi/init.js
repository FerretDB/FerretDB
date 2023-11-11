function initCodapi() {
    document.querySelectorAll(".codapi-snippet").forEach((el) => {
        const snippet = document.createElement("codapi-snippet");
        setAttribute(snippet, el, "sandbox");
        setAttribute(snippet, el, "editor");
        setAttribute(snippet, el, "template");
        el.replaceWith(snippet);
    });
}

function setAttribute(dst, src, attrName) {
    if (!src.hasAttribute(attrName)) {
        return;
    }
    dst.setAttribute(attrName, src.getAttribute(attrName));
}

addEventListener("load", initCodapi);
