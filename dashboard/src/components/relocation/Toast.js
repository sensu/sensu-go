import React from "react";
import PropTypes from "prop-types";
import classNames from "classnames";
import CheckCircleIcon from "@material-ui/icons/CheckCircle";
import ErrorIcon from "@material-ui/icons/Error";
import InfoIcon from "@material-ui/icons/Info";
import CloseIcon from "@material-ui/icons/Close";
import green from "@material-ui/core/colors/green";
import amber from "@material-ui/core/colors/amber";
import IconButton from "@material-ui/core/IconButton";
import SnackbarContent from "@material-ui/core/SnackbarContent";
import WarningIcon from "@material-ui/icons/Warning";
import { withStyles } from "@material-ui/core/styles";

import uniqueId from "/utils/uniqueId";
import Timer from "/components/util/Timer";
import CircularProgress from "/components/partials/CircularProgress";

const icons = {
  success: CheckCircleIcon,
  warning: WarningIcon,
  error: ErrorIcon,
  info: InfoIcon,
};

const styles = theme => ({
  message: {
    display: "flex",
    alignItems: "center",

    [theme.breakpoints.down("md")]: {
      paddingLeft: "env(safe-area-inset-left)",
    },
  },
  action: {
    [theme.breakpoints.down("md")]: {
      paddingRight: "env(safe-area-inset-right)",
    },
  },
  success: {
    backgroundColor: green[600],
  },
  error: {
    backgroundColor: theme.palette.error.dark,
  },
  info: {
    backgroundColor: theme.palette.primary.dark,
  },
  warning: {
    backgroundColor: amber[700],
  },
  icon: {
    fontSize: 20,
  },
  variantIcon: {
    opacity: 0.9,
    fontSize: 20,
    marginRight: theme.spacing.unit,
  },
});

class Toast extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    message: PropTypes.node,
    variant: PropTypes.oneOf(Object.keys(icons)),
    onClose: PropTypes.func,
    maxAge: PropTypes.number,
  };

  static defaultProps = {
    maxAge: 0,
    onClose: () => {},
    variant: undefined,
    message: undefined,
  };

  id = `Toast-${uniqueId()}`;

  render() {
    const { classes, message, onClose, variant, maxAge } = this.props;
    const Icon = icons[variant];

    const messageId = `${this.id}-message`;

    const closeButton = (
      <IconButton
        key="close"
        aria-label="Close"
        color="inherit"
        className={classes.close}
        onClick={onClose}
      >
        <CloseIcon className={classes.icon} />
      </IconButton>
    );

    return (
      <SnackbarContent
        className={classNames(classes.root, classes[variant])}
        classes={{ action: classes.action, message: classes.message }}
        aria-describedby={messageId}
        message={
          <span id={messageId}>
            {Icon && <Icon className={classes.variantIcon} />}
            {message}
          </span>
        }
        action={[
          maxAge !== 0 ? (
            <Timer key={closeButton.props.key} delay={maxAge} onEnd={onClose}>
              {progress => (
                <CircularProgress width={4} value={progress} opacity={0.5}>
                  {closeButton}
                </CircularProgress>
              )}
            </Timer>
          ) : (
            closeButton
          ),
        ]}
      />
    );
  }
}

export default withStyles(styles)(Toast);
