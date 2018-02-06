import { createMuiTheme } from "material-ui/styles";
import { getContrastRatio } from "material-ui/styles/colorManipulator";
import {
  light as lightTheme,
  dark as darkTheme,
} from "material-ui/styles/createPalette";

import colors from "../../colors";

const DefaultTheme = createMuiTheme({
  palette: {
    // type: "dark",

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

    // HACK: Reduce threshold white text while testing green theme
    getContrastText: color => {
      if (getContrastRatio(color, colors.common.black) < 10) {
        return darkTheme.text.primary;
      }
      return lightTheme.text.primary;
    },
  },
});

export default DefaultTheme;
