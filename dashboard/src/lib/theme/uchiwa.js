import createTheme from "/lib/theme/createTheme";
import colors from "/lib/colors";

const theme = (type = "light") =>
  createTheme({
    palette: {
      type,

      primary: {
        light: colors.cornflowerBlue[300],
        main: colors.cornflowerBlue[500],
        dark: colors.cornflowerBlue[600],
        contrastText: "#F3F5F7",
      },
      secondary: {
        light: colors.appleGreen[300],
        main: colors.appleGreen[500],
        dark: colors.appleGreen[700],
        contrastText: "#1D2237",
      },
    },
  });

export default theme;
