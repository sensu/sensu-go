import React from "react";
import PropTypes from "prop-types";

import { withStyles } from "material-ui/styles";
import Paper from "material-ui/Paper";
import Button from "material-ui/Button";
import TextField from "material-ui/TextField";
import DefaultThemeProvider, { ExteriorTheme } from "./Theme";
import { authenticate } from "../utils/authentication";

// Temp.
const styles = theme => ({
  "@global": {
    html: {
      background: theme.palette.primary["400"],
      WebkitFontSmoothing: "antialiased", // Antialiasing.
      MozOsxFontSmoothing: "grayscale", // Antialiasing.
      boxSizing: "border-box",
    },
    "*, *:before, *:after": {
      boxSizing: "inherit",
    },
    body: {
      margin: 0,
    },
  },
  container: {
    display: "flex",
    flexWrap: "wrap",
  },
  loginCard: {
    margin: "150px auto",
    padding: "40px 28px",
    width: 300,
    textAlign: "right",
  },
  textField: {
    marginBottom: theme.spacing.unit,
  },
  button: {
    textTransform: "uppercase",
  },
});

class Login extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
  };

  constructor(props) {
    super(props);

    this.state = {};
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleSubmit(ev) {
    const { username, password } = this.state;

    // TODO: Redirect to dashboard.
    const handleSuccess = () => this.setState({ disabled: false });
    const handleFailure = () => {
      this.setState({
        disabled: false,
        authError: "Bad username or password.",
      });
    };

    // Stop default form behaviour
    ev.preventDefault();

    // Disable form
    this.setState({ disabled: true });

    // Authenticate user
    const authPromise = authenticate(username, password);
    authPromise.then(handleSuccess).catch(handleFailure);
  }

  render() {
    const { classes } = this.props;
    const { authError, disabled } = this.state;

    const changeField = (name, val) =>
      this.setState({
        authError: null,
        [name]: val,
      });
    const handleUsername = ev => changeField("username", ev.target.value);
    const handlePassword = ev => this.setState({ password: ev.target.value });

    return (
      <DefaultThemeProvider theme={ExteriorTheme}>
        <Paper className={classes.loginCard}>
          <form onSubmit={this.handleSubmit}>
            <TextField
              name="username"
              label="Username"
              aria-label="Username"
              autoComplete="username"
              spellCheck="false"
              className={classes.textField}
              fullWidth
              margin="normal"
              onChange={handleUsername}
              disabled={disabled}
              error={!!authError}
            />
            <TextField
              type="password"
              name="password"
              label="Password"
              aria-label="Password"
              className={classes.textField}
              fullWidth
              onChange={handlePassword}
              disabled={disabled}
              error={!!authError}
              helperText={authError}
            />
            <Button
              type="submit"
              color="primary"
              raised
              disabled={disabled}
              className={classes.button}
            >
              Log in
            </Button>
          </form>
        </Paper>
      </DefaultThemeProvider>
    );
  }
}

export default withStyles(styles)(Login);
