import createTheme from "/lib/theme/createTheme";
import colors from "/lib/colors";

const theme = (type = "dark") =>
  createTheme({
    palette: {
      type,

      primary: {
        light: colors.paynesGrey[300],
        main: colors.paynesGrey[500],
        dark: colors.paynesGrey[600],
        contrastText: "#F3F5F7",
      },
      secondary: {
        light: colors.pistachio[300],
        main: colors.pistachio[500],
        dark: colors.pistachio[700],
        contrastText: "#1D2237",
      },
    },
  });

export default theme;
