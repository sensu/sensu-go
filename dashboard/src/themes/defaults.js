// Overrides defaults provided by Material-UI (material-ui/src/styles)
const defaults = {
  direction: "ltr",
  palette: {
    success: "rgb(151, 198, 115)",
    warning: "rgb(213, 171, 59)",
    critical: "rgb(204, 62, 100)",
    unknown: "rgb(211, 97, 53)",
  },
  typography: () => {
    // Prefer 'Display' for larger fonts
    // https://developer.apple.com/ios/human-interface-guidelines/visual-design/typography/
    const title = `"SF Pro Display", "Roboto", "Helvetica", "Arial", sans-serif`;
    const body = `"SF Pro Text", "Roboto", "Helvetica", "Arial", sans-serif`;
    const mono = `"SFMono-Regular", Consolas, "Liberation Mono", Menlo,Courier, monospace`;

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
      monospace: {
        fontFamily: mono,
      },
    };
  },
  overrides: {
    MuiDialog: {
      root: {
        backdropFilter: "blur(3px)",
        transition: "backdropFilter 300ms ease-in-out",
      },
    },
  },
};

export default defaults;
