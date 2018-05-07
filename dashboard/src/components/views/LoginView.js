import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";
import { compose } from "recompose";
import { withApollo } from "react-apollo";

import Paper from "material-ui/Paper";
import Button from "material-ui/Button";
import TextField from "material-ui/TextField";
import Typography from "material-ui/Typography";

import createTokens from "/mutations/createTokens";

import Logo from "/icons/SensuLogoGraphic";
import Wordmark from "/icons/SensuWordmark";

class LoginView extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    client: PropTypes.object.isRequired,
  };

  static styles = theme => ({
    "@global html": {
      background: theme.palette.background.main,
    },
    loginCard: {
      margin: "0 auto",
      padding: 48,
      maxWidth: 450,
      textAlign: "left",
      alignSelf: "center",
      height: "auto",
      minHeight: 500,
    },
    textField: {
      marginBottom: theme.spacing.unit,
    },
    actionsContainer: {
      marginTop: theme.spacing.unit * 5,
      textAlign: "right",
    },
    button: {
      textTransform: "uppercase",
    },
    icon: {
      color: theme.palette.secondary.main,
      marginBottom: theme.spacing.unit,
    },
    wordmark: {
      color: theme.palette.secondary.main,
      fontSize: theme.spacing.unit * 1.5,
      marginBottom: theme.spacing.unit,
      marginLeft: theme.spacing.unit / 2,
    },
    headline: {
      marginTop: theme.spacing.unit * 2,
    },
    form: {
      marginTop: theme.spacing.unit * 3,
    },
    root: {
      display: "flex",
      alignItems: "stretch",
      minHeight: "100vh",
      width: "100%",
    },
  });

  state = {
    username: null,
    password: null,
    authError: null,
  };

  handleSubmit = ev => {
    const { client } = this.props;
    const { username, password } = this.state;

    // Stop default form behaviour
    ev.preventDefault();

    // Disable form
    this.setState({ disabled: true });

    createTokens(client, { username, password }).catch(error => {
      // eslint-disable-next-line no-console
      console.error(error);
      this.setState({
        disabled: false,
        authError: "Bad username or password.",
      });
    });
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
      <div className={classes.root}>
        <Paper className={classes.loginCard}>
          <Logo className={classes.icon} />
          <Wordmark className={classes.wordmark} />
          <div className={classes.headline}>
            <Typography variant="headline">Sign in</Typography>
            <Typography variant="subheading">
              with your Sensu Account
            </Typography>
          </div>
          <form className={classes.form} onSubmit={this.handleSubmit}>
            <TextField
              name="username"
              label="Username"
              aria-label="Username"
              autoComplete="username"
              autoCorrect="false"
              autoCapitalize="none"
              spellCheck={false}
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
              autoComplete="current-password"
              className={classes.textField}
              fullWidth
              onChange={handlePassword}
              disabled={disabled}
              error={!!authError}
              helperText={authError}
            />
            <div className={classes.actionsContainer}>
              <Button
                type="submit"
                color="primary"
                variant="raised"
                disabled={disabled}
                className={classes.button}
              >
                Log in
              </Button>
            </div>
          </form>
        </Paper>
      </div>
    );
  }
}

export default compose(withApollo, withStyles(LoginView.styles))(LoginView);
