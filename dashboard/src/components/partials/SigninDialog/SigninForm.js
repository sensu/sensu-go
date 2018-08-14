import React from "react";
import PropTypes from "prop-types";
import createStyledComponent from "/components/createStyledComponent";

import Button from "@material-ui/core/Button";
import DialogActions from "@material-ui/core/DialogActions";
import TextField from "@material-ui/core/TextField";

const StyledDialoagActions = createStyledComponent({
  name: "SignInForm.Actions",
  component: DialogActions,
  styles: theme => ({
    marginTop: theme.spacing.unit * 5,
    textAlign: "right",
  }),
});

class SignInForm extends React.Component {
  static propTypes = {
    disabled: PropTypes.bool,
    error: PropTypes.string,
    onSubmit: PropTypes.func.isRequired,
  };

  static defaultProps = {
    disabled: false,
    error: null,
  };

  handleSubmit = ev => {
    const { username, password } = this.state;

    ev.preventDefault();
    this.props.onSubmit({ username, password });
  };

  render() {
    const { disabled, error } = this.props;

    const changeField = (name, val) =>
      this.setState({
        authError: null,
        [name]: val,
      });
    const handleUsername = ev => changeField("username", ev.target.value);
    const handlePassword = ev => this.setState({ password: ev.target.value });

    return (
      <form onSubmit={this.handleSubmit}>
        <TextField
          name="username"
          label="Username"
          aria-label="Username"
          autoComplete="username"
          autoCorrect="false"
          autoCapitalize="none"
          spellCheck={false}
          fullWidth
          margin="normal"
          onChange={handleUsername}
          disabled={disabled}
          error={!!error}
        />
        <TextField
          type="password"
          name="password"
          label="Password"
          aria-label="Password"
          autoComplete="current-password"
          fullWidth
          onChange={handlePassword}
          disabled={disabled}
          error={!!error}
          helperText={error}
        />
        <StyledDialoagActions>
          <Button
            type="submit"
            color="primary"
            variant="raised"
            disabled={disabled}
          >
            Sign in
          </Button>
        </StyledDialoagActions>
      </form>
    );
  }
}

export default SignInForm;
