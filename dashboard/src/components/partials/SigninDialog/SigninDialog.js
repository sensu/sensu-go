import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import { compose, withProps } from "recompose";
import { withApollo } from "react-apollo";
import { when } from "/utils/promise";
import { UnauthorizedError } from "/errors/FetchError";
import withMobileDialog from "@material-ui/core/withMobileDialog";
import createTokens from "/mutations/createTokens";
import createStyledComponent from "/components/createStyledComponent";

import Dialog from "@material-ui/core/Dialog";
import LinearProgress from "@material-ui/core/LinearProgress";
import Logo from "/icons/SensuLogo";
import Slide from "@material-ui/core/Slide";
import Typography from "@material-ui/core/Typography";

import Form from "./SigninForm";

const Branding = createStyledComponent({
  component: Typography,
  styles: theme => ({
    marginBottom: theme.spacing.unit * 3,

    "& svg": {
      marginRight: theme.spacing.unit / 2,
    },
  }),
});

const Headline = createStyledComponent({
  component: Typography,
  styles: theme => ({
    marginBottom: theme.spacing.unit * 3,
  }),
});

const styles = theme => ({
  root: {
    margin: "0 auto",
    padding: 48,
    maxWidth: 450,
    textAlign: "left",
    alignSelf: "center",
    height: "auto",
    minHeight: 500,
  },
});

class SignInView extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    client: PropTypes.object.isRequired,
    fullScreen: PropTypes.bool,
    TransitionComponent: PropTypes.func,
    onSuccess: PropTypes.func,
    open: PropTypes.bool,
  };

  static defaultProps = {
    fullScreen: false,
    TransitionComponent: withProps({ direction: "up" })(Slide),
    onSuccess: () => {},
    open: true,
  };

  state = {
    authError: null,
    loading: false,
  };

  handleSubmit = ({ username, password }) => {
    this.setState({ loading: true });

    createTokens(this.props.client, { username, password })
      .then(this.props.onSuccess)
      .catch(
        when(UnauthorizedError, () => {
          this.setState({
            loading: false,
            authError: "Bad username or password.",
          });
          // TODO: Handle other fetch error cases to show an inline error state
          // instead of triggering global error modal.
        }),
      );
  };

  render() {
    const {
      classes,
      fullScreen,
      TransitionComponent,
      onSuccess,
      open,
      ...props
    } = this.props;
    const { authError, loading } = this.state;

    return (
      <Dialog
        fullScreen={fullScreen}
        open={open}
        TransitionComponent={TransitionComponent}
        {...props}
      >
        <LinearProgress
          variant={loading ? "indeterminate" : "determinate"}
          value={0}
        />
        <div className={classes.root}>
          <Branding color="secondary">
            <Logo />
          </Branding>
          <Headline>
            <Typography variant="headline">Sign in</Typography>
            <Typography variant="subheading">
              with your Sensu Account
            </Typography>
          </Headline>
          <Form
            disabled={loading}
            error={authError}
            onSubmit={this.handleSubmit}
          />
        </div>
      </Dialog>
    );
  }
}

export default compose(
  withApollo,
  withStyles(styles),
  withMobileDialog({ breakpoint: "xs" }),
)(SignInView);
