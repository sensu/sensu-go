// Overrides defaults provided by Material-UI (material-ui/src/styles)
const defaults = {
  direction: "ltr",
  palette: {
    // ...
  },
  typography: () => {
    const body = `"SF Pro Text", "Roboto", "Helvetica", "Arial", sans-serif`;
    const title = `"SF Pro Display", "Roboto", "Helvetica", "Arial", sans-serif`;

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
