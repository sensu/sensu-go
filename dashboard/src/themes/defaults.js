// Overrides defaults provided by Material-UI (material-ui/src/styles)
const defaults = {
  direction: "ltr",
  palette: {
    // ...
  },
  typography: () => {
    // Prefer 'Display' for larger fonts
    // https://developer.apple.com/ios/human-interface-guidelines/visual-design/typography/
    const title = `"SF Pro Display", "Roboto", "Helvetica", "Arial", sans-serif`;
    const body = `"SF Pro Text", "Roboto", "Helvetica", "Arial", sans-serif`;

    return {
      fontFamily: title,
      body1: {
        fontFamily: body,
      },
      body2: {
        fontFamily: body,
      },
      button: {
        fontFamily: title,
        fontWeight: "bold",
      },
    };
  },
};

export default defaults;
