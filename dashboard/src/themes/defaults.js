// Overrides defaults provided by Material-UI (material-ui/src/styles)
const defaults = {
  direction: "ltr",
  palette: {
    // ...
  },
  typography: () => {
    const fontFamily = `"SF Pro Text", "Roboto", "Helvetica", "Arial", sans-serif`;

    return {
      fontFamily,
      caption: {
        fontFamily: `"SF Pro Display", "Roboto", "Helvetica", "Arial", sans-serif`,
      },
    };
  },
};

export default defaults;
