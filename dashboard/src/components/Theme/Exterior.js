import { createMuiTheme } from "material-ui/styles";
import colors from "../../colors";

const Theme = createMuiTheme({
  palette: {
    primary: colors.magenta,
    secondary: colors.slateBlue,
  },
});

export default Theme;
