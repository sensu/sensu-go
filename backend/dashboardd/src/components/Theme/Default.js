import { createMuiTheme } from "material-ui/styles";
import colors from "../../colors";

const DefaultTheme = createMuiTheme({
  palette: {
    primary: colors.indigo,
    secondary: colors.salmon,
  },
});

export default DefaultTheme;
