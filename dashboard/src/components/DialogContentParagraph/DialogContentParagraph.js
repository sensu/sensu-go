import { withStyles } from "@material-ui/core/styles";
import DialogContentText from "@material-ui/core/DialogContentText";

const StyledDialogContentText = withStyles(() => ({
  root: { marginBottom: "2rem" },
}))(DialogContentText);

export default StyledDialogContentText;
