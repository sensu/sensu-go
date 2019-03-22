export default {
  [[
    "a",
    "abbr",
    "acronym",
    "address",
    "applet",
    "article",
    "aside",
    "audio",
    "b",
    "big",
    "blockquote",
    "body",
    "canvas",
    "caption",
    "center",
    "cite",
    "code",
    "dd",
    "del",
    "details",
    "dfn",
    "div",
    "dl",
    "dt",
    "em",
    "embed",
    "fieldset",
    "figcaption",
    "figure",
    "footer",
    "form",
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "header",
    "hgroup",
    "html",
    "i",
    "iframe",
    "img",
    "ins",
    "kbd",
    "label",
    "legend",
    "li",
    "mark",
    "menu",
    "nav",
    "object",
    "ol",
    "output",
    "p",
    "pre",
    "q",
    "ruby",
    "s",
    "samp",
    "section",
    "small",
    "span",
    "strike",
    "strong",
    "sub",
    "summary",
    "sup",
    "table",
    "tbody",
    "td",
    "tfoot",
    "th",
    "thead",
    "time",
    "tr",
    "tt",
    "u",
    "ul",
    "var",
    "video",
  ].join(",")]: {
    margin: 0,
    padding: 0,
    border: 0,
    font: "inherit",
    verticalAlign: "baseline",
  },

  // Style "unknown" block tags in IE11.
  [[
    "article",
    "aside",
    "canvas",
    "figcaption",
    "figure",
    "footer",
    "header",
    "hgroup",
    "main",
    "output",
    "section",
  ].join(",")]: {
    display: "block",
  },

  body: {
    lineHeight: 1,
  },

  "ol, ul": {
    listStyle: "none",
  },

  "blockquote, q": {
    quotes: "none",
    "&::before, &::after": {
      content: '""',
      display: "none",
    },
  },

  table: {
    borderCollapse: "collapse",
    borderSpacing: 0,
  },

  "input, textarea, button, select": {
    appearance: "none",
    background: "none",
    borderColor: "currentColor",
    borderStyle: "none",
    borderWidth: "medium",
    borderRadius: 0,
    margin: 0,
    padding: 0,
    color: "inherit",
    fontStyle: "inherit",
    fontVariant: "inherit",
    fontWeight: "inherit",
    fontStretch: "inherit",
    fontSize: "inherit",
    fontFamily: "inherit",
    lineHeight: "inherit",
    verticalAlign: "baseline",
  },

  "::placeholder": {
    opacity: 1,
    color: "inherit",
    fontStyle: "inherit",
    fontVariant: "inherit",
    fontWeight: "inherit",
    fontStretch: "inherit",
    fontSize: "inherit",
    fontFamily: "inherit",
  },

  // Reset anchor elements.
  a: {
    textDecoration: "none",

    // Remove the gray background on active links in IE 10.
    backgroundColor: "transparent",

    // Remove gaps in links underline in iOS 8+ and Safari 8+.
    "-webkit-text-decoration-skip": "objects",
  },

  // Remove input overlay elements.
  "::-webkit-inner-spin-button": {
    appearance: "none",
  },
  "::-webkit-outer-spin-button": {
    appearance: "none",
  },
  "::-webkit-search-cancel-button": {
    appearance: "none",
  },
  "::-webkit-search-results-button": {
    appearance: "none",
  },
  "::-ms-clear, ::-ms-reveal": {
    display: "none",
  },
  "::-webkit-contacts-auto-fill-button": {
    display: "none !important",
  },

  input: {
    // Remove firefox number input arrow buttons.
    "-moz-appearance": "textfield",
  },

  hr: {
    // Correct box-sizing in FF
    boxSizing: "content-box",

    // Show overflow in IE.
    height: 0,
    overflow: "visible",
  },

  img: {
    // Remove border style from linked images on IE.
    borderStyle: "none",
  },

  "svg:not(:root)": {
    // Hide the overflow in IE.
    overflow: "hidden",
  },

  "*, *:before, *:after": {
    // Add border box sizing in all browsers
    boxSizing: "inherit",

    // Remove repeating backgrounds in all browsers
    backgroundRepeat: "no-repeat",

    // Add text decoration inheritance in all browsers
    textDecoration: "inherit",

    // Add vertical alignment inheritence in all browsers
    verticalAlign: "inherit",
  },

  html: {
    // Add border box sizing in all browsers
    boxSizing: "border-box",

    // Add the default cursor in all browsers
    cursor: "default",

    // Prevent font size adjustments after orientation changes in IE and iOS.
    "-webkit-text-size-adjust": "100%",
    "-ms-text-size-adjust": "100%",
  },
};
