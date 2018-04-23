import createTheme from "/themes/createTheme";
import colors from "/colors";

const theme = (type = "dark") =>
  createTheme({
    palette: {
      type,

      primary: {
        light: colors.magenta[300],
        main: colors.magenta[500],
        dark: colors.magenta[600],
      },
      secondary: {
        light: colors.slateBlue[300],
        main: colors.slateBlue[500],
        dark: colors.slateBlue[700],
      },
    },
  });

export default theme;
