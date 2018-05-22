import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";

import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/ButtonBase";

const styles = theme => ({
  root: {
    height: theme.spacing.unit * 3,
    paddingTop: 12,
    paddingBottom: 12,
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
    boxSizing: "content-box",
  },
});

export class TableListButton extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
  };

  static defaultProps = {
    className: "",
  };

  render() {
    const { classes, className: classNameProp, ...props } = this.props;
    const className = classnames(classes.root, classNameProp);

    return <Button className={className} {...props} />;
  }
}

export default withStyles(styles)(TableListButton);
