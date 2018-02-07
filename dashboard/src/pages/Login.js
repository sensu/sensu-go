import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape } from "found";
import { withStyles } from "material-ui/styles";

import Paper from "material-ui/Paper";
import Button from "material-ui/Button";
import TextField from "material-ui/TextField";
import ExteriorWrapper from "../components/ExteriorWrapper";
import { authenticate } from "../utils/authentication";

// defaultRoute describe the location where the user will land after a
// successful login.
const defaultRoute = "/";

class Login extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    router: routerShape.isRequired,
  };

  // Temp.
  static styles = theme => ({
    loginCard: {
      margin: "0 auto",
      padding: "40px 28px",
      width: 300,
      textAlign: "right",
      alignSelf: "center",
    },
    textField: {
      marginBottom: theme.spacing.unit,
    },
    button: {
      textTransform: "uppercase",
    },
  });

  state = {
    username: null,
    password: null,
  };

  handleSubmit = ev => {
    const { username, password } = this.state;
    const { router } = this.props;

    const handleSuccess = () => router.replace(defaultRoute);
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
  };

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
      <ExteriorWrapper>
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
      </ExteriorWrapper>
    );
  }
}

export default compose(withStyles(Login.styles), withRouter)(Login);
