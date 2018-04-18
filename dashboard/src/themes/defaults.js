// Overrides defaults provided by Material-UI (material-ui/src/styles)
const defaults = {
  direction: "ltr",
  palette: {
    green: "rgb(151, 198, 115)",
    yellow: "rgb(213, 171, 59)",
    orange: "rgb(211, 97, 53)",
    red: "rgb(204, 62, 100)",
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
