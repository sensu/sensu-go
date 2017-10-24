import { createMuiTheme } from "material-ui/styles";
import { getContrastRatio } from "material-ui/styles/colorManipulator";
import {
  light as lightTheme,
  dark as darkTheme,
} from "material-ui/styles/createPalette";

import colors from "../../colors";

const DefaultTheme = createMuiTheme({
  palette: {
    primary: colors.slateBlue,
    secondary: colors.red,

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
