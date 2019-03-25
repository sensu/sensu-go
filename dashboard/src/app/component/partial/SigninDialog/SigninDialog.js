import React from "react";
import PropTypes from "prop-types";
import { compose, withProps } from "recompose";
import { withApollo } from "react-apollo";

import withMobileDialog from "@material-ui/core/withMobileDialog";
import Dialog from "@material-ui/core/Dialog";
import DialogContent from "@material-ui/core/DialogContent";
import LinearProgress from "@material-ui/core/LinearProgress";
import Slide from "@material-ui/core/Slide";
import Typography from "@material-ui/core/Typography";

import { when } from "/lib/util/promise";
import { UnauthorizedError } from "/lib/error/FetchError";
import createTokens from "/lib/mutation/createTokens";
import createStyledComponent from "/lib/component/util/createStyledComponent";
import Logo from "/lib/component/icon/SensuLogo";

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
  component: "div",
  styles: theme => ({
    marginBottom: theme.spacing.unit * 3,
  }),
});

const Container = createStyledComponent({
  component: DialogContent,
  styles: theme => ({
    padding: theme.spacing.unit * 6,
    height: "auto",
    minHeight: 500,
  }),
});

class SignInView extends React.Component {
  static propTypes = {
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
        <Container>
          <Branding color="secondary">
            <Logo />
          </Branding>
          <Headline>
            <Typography variant="h5">Sign in</Typography>
            <Typography variant="subtitle1">with your Sensu Account</Typography>
          </Headline>
          <Form
            disabled={loading}
            error={authError}
            onSubmit={this.handleSubmit}
          />
        </Container>
      </Dialog>
    );
  }
}

const enhance = compose(
  withApollo,
  withMobileDialog({ breakpoint: "xs" }),
);
export default enhance(SignInView);
