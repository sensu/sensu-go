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
import WarningIcon from "@material-ui/icons/Warning";
import { withStyles } from "@material-ui/core/styles";
import { emphasize } from "@material-ui/core/styles/colorManipulator";
import Paper from "@material-ui/core/Paper";
import Typography from "@material-ui/core/Typography";

import uniqueId from "/utils/uniqueId";
import Timer from "/components/util/Timer";
import CircularProgress from "/components/partials/CircularProgress";

const icons = {
  success: CheckCircleIcon,
  warning: WarningIcon,
  error: ErrorIcon,
  info: InfoIcon,
};

export const styles = theme => {
  const emphasis = theme.palette.type === "light" ? 0.8 : 0.98;
  const backgroundColor = emphasize(theme.palette.background.default, emphasis);

  return {
    /* Styles applied to the root element. */
    root: {
      position: "relative",
      overflow: "hidden",
      color: theme.palette.getContrastText(backgroundColor),
      backgroundColor,
      display: "flex",
      alignItems: "center",
      [theme.breakpoints.up("md")]: {
        minWidth: 288,
        maxWidth: 568,
        borderRadius: theme.shape.borderRadius,
      },
      [theme.breakpoints.down("sm")]: {
        flexGrow: 1,
      },
    },
    progress: {
      position: "absolute",
      top: 0,
      left: 0,
      right: 0,
    },

    /* Styles applied to the message wrapper element. */
    message: {
      paddingTop: 14,
      paddingBottom: 14,
      paddingLeft: 24,

      display: "flex",
      alignItems: "center",

      [theme.breakpoints.down("md")]: {
        marginLeft: "env(safe-area-inset-left)",
      },

      "& strong": {
        fontWeight: 600,
      },
    },
    /* Styles applied to the action wrapper element if `action` is provided. */
    action: {
      display: "flex",
      alignItems: "center",
      marginLeft: "auto",
      paddingTop: 6,
      paddingBottom: 6,
      paddingLeft: 24,
      paddingRight: 16,
      marginRight: -8,

      [theme.breakpoints.down("md")]: {
        marginRight: "env(safe-area-inset-right)",
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
  };
};

class Toast extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    message: PropTypes.node,
    variant: PropTypes.oneOf(Object.keys(icons)),
    onClose: PropTypes.func.isRequired,
    maxAge: PropTypes.number,
    progress: PropTypes.node,
  };

  static defaultProps = {
    maxAge: 0,
    variant: undefined,
    message: undefined,
    progress: undefined,
  };

  id = `Toast-${uniqueId()}`;

  render() {
    const {
      classes,
      message,
      onClose,
      variant,
      maxAge,
      progress: progressBar,
    } = this.props;
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
      <Paper
        component={Typography}
        headlineMapping={{
          body1: "div",
        }}
        role="alertdialog"
        square
        elevation={6}
        className={classNames(classes.root, classes[variant])}
        aria-describedby={messageId}
      >
        <div className={classes.progress}>{progressBar}</div>
        <div id={messageId} className={classes.message}>
          {Icon && <Icon className={classes.variantIcon} />}
          {message}
        </div>
        <div className={classes.action}>
          {maxAge !== 0 ? (
            <Timer key={closeButton.props.key} delay={maxAge} onEnd={onClose}>
              {progress => (
                <CircularProgress width={4} value={progress} opacity={0.5}>
                  {closeButton}
                </CircularProgress>
              )}
            </Timer>
          ) : (
            closeButton
          )}
        </div>
      </Paper>
    );
  }
}

export default withStyles(styles)(Toast);
