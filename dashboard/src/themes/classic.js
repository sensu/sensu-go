import createTheme from "/themes/createTheme";
import colors from "/colors";

const theme = (type = "light") =>
  createTheme({
    palette: {
      type,

      primary: {
        light: colors.appleGreen[300],
        main: colors.appleGreen[500],
        dark: colors.appleGreen[600],
      },
      secondary: {
        light: colors.pistachio[300],
        main: colors.pistachio[500],
        dark: colors.pistachio[700],
      },
      contrastThreshold: 1,
    },
  });

export default theme;
