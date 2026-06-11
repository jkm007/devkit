/**
 * Sanitize HTML string to prevent XSS attacks.
 *
 * Uses the browser's DOMParser to parse HTML and removes all potentially
 * dangerous elements and attributes:
 * - Script tags and other code-executing elements
 * - Event handler attributes (onclick, onload, etc.)
 * - Dangerous URL schemes (javascript:, data:) in href/src attributes
 * - Malicious CSS in style attributes (expression(), behavior:)
 *
 * This is a lightweight alternative to DOMPurify for sanitizing simple
 * HTML content such as rich text snippets.
 */

/** Elements that are always removed regardless of context. */
const FORBIDDEN_TAGS = new Set([
  'applet',
  'base',
  'button',
  'embed',
  'form',
  'iframe',
  'input',
  'link',
  'meta',
  'object',
  'script',
  'select',
  'style',
  'textarea',
]);

/** Attributes that are always removed (event handlers, etc.). */
const FORBIDDEN_ATTRS = new Set([
  'onabort',
  'onanimationend',
  'onanimationiteration',
  'onanimationstart',
  'onbeforeunload',
  'onblur',
  'onchange',
  'onclick',
  'onclose',
  'ondblclick',
  'onerror',
  'onfocus',
  'onhashchange',
  'onkeydown',
  'onkeypress',
  'onkeyup',
  'onload',
  'onmessage',
  'onmousedown',
  'onmousemove',
  'onmouseout',
  'onmouseover',
  'onmouseup',
  'onopen',
  'onpagehide',
  'onpageshow',
  'onpopstate',
  'onreset',
  'onresize',
  'onscroll',
  'onselect',
  'onstorage',
  'onsubmit',
  'ontouchcancel',
  'ontouchend',
  'ontouchmove',
  'ontouchstart',
  'ontransitionend',
  'onunload',
]);

/** URL-like attributes that must not contain javascript: or data: schemes. */
const URL_ATTRS = new Set([
  'action',
  'formaction',
  'href',
  'src',
  'xlink:href',
]);

/** Allowed URL schemes. Only http(s), mailto, tel, relative paths and anchors are safe. */
const SAFE_URL_PATTERN = /^(https?|mailto|tel|#|\/)/i;

/**
 * Sanitize an HTML string by removing dangerous elements and attributes.
 *
 * @param html - The raw HTML string to sanitize.
 * @returns A sanitized HTML string safe for insertion via v-html.
 */
export function sanitizeHtml(html: string): string {
  if (!html) return '';

  const parser = new DOMParser();
  const doc = parser.parseFromString(html, 'text/html');
  sanitizeNode(doc.body);
  return doc.body.innerHTML;
}

/**
 * Recursively sanitize a DOM node tree, removing forbidden elements
 * and attributes in-place.
 */
function sanitizeNode(node: Node): void {
  // Iterate in reverse because we may remove children during traversal
  const childNodes = Array.from(node.childNodes);

  for (const child of childNodes) {
    if (child.nodeType === Node.ELEMENT_NODE) {
      const el = child as Element;
      const tagName = el.tagName.toLowerCase();

      // Remove forbidden elements entirely (e.g., <script>, <iframe>)
      if (FORBIDDEN_TAGS.has(tagName)) {
        node.removeChild(child);
        continue;
      }

      // Scrub forbidden attributes from the element
      for (const attr of Array.from(el.attributes)) {
        const attrName = attr.name.toLowerCase();

        // Remove event handler attributes (onclick, onload, etc.)
        if (FORBIDDEN_ATTRS.has(attrName)) {
          el.removeAttribute(attr.name);
          continue;
        }

        // Check URL attributes for dangerous schemes
        if (
          URL_ATTRS.has(attrName) &&
          !SAFE_URL_PATTERN.test(attr.value.trim())
        ) {
          el.removeAttribute(attr.name);
          continue;
        }

        // Strip style attributes containing CSS injection vectors
        if (attrName === 'style') {
          const value = attr.value.toLowerCase();
          if (
            value.includes('expression(') ||
            value.includes('javascript:') ||
            value.includes('behavior:') ||
            value.includes('url(')
          ) {
            el.removeAttribute(attr.name);
          }
        }
      }

      // Recurse into child elements
      sanitizeNode(el);
    } else if (child.nodeType === Node.COMMENT_NODE) {
      // Remove HTML comments (could contain conditional comments)
      node.removeChild(child);
    }
  }
}
