import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

const styles = theme => {
  const toolbar = theme.mixins.toolbar;
  const xsBrk = `${theme.breakpoints.up("xs")} and (orientation: landscape)`;
  const smBrk = theme.breakpoints.up("sm");
  const calcTopWithFallback = size => ({
    top: `calc(${size}px + env(safe-area-inset-top))`,
    fallbacks: [{ top: size }],
  });

  return {
    root: {
      padding: `0 ${theme.spacing.unit * 2}px`,
      backgroundColor: theme.palette.primary.light,
      color: theme.palette.primary.contrastText,
      display: "flex",
      alignItems: "center",
      height: 56,
      zIndex: theme.zIndex.appBar - 1,
      "& *": {
        color: theme.palette.primary.contrastText,
      },
    },
    active: {
      backgroundColor: theme.palette.primary.main,
    },
    sticky: {
      position: "sticky",
      ...calcTopWithFallback(toolbar.minHeight),
      [xsBrk]: {
        ...calcTopWithFallback(toolbar[xsBrk].minHeight),
      },
      [smBrk]: {
        ...calcTopWithFallback(toolbar[smBrk].minHeight),
      },
      color: theme.palette.primary.contrastText,
    },
  };
};

export class TableListHeader extends React.Component {
  static propTypes = {
    active: PropTypes.bool,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
    sticky: PropTypes.bool,
  };

  static defaultProps = {
    active: false,
    sticky: false,
    className: "",
  };

  render() {
    const {
      active,
      sticky,
      classes,
      className: classNameProp,
      children,
    } = this.props;

    const className = classnames(classes.root, classNameProp, {
      [classes.active]: active,
      [classes.sticky]: sticky,
    });

    return (
      <Typography component="div" className={className}>
        {children}
      </Typography>
    );
  }
}

export default withStyles(styles)(TableListHeader);
