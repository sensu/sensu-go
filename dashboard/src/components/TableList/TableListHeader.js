import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

const styles = theme => ({
  root: {
    padding: `0 ${theme.spacing.unit * 2}px`,
    backgroundColor: theme.palette.primary.light,
    color: theme.palette.primary.contrastText,
    display: "flex",
    alignItems: "center",
    height: 56,
    zIndex: theme.zIndex.appBar - 1,
  },
  active: {
    backgroundColor: theme.palette.primary.main,
  },
});

export class TableListHeader extends React.Component {
  static propTypes = {
    active: PropTypes.bool,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
  };

  static defaultProps = {
    active: false,
    className: "",
  };

  render() {
    const { active, classes, className: classNameProp, children } = this.props;
    const className = classnames(classes.root, classNameProp, {
      [classes.active]: active,
    });

    return (
      <Typography component="div" className={className}>
        {children}
      </Typography>
    );
  }
}

export default withStyles(styles)(TableListHeader);
