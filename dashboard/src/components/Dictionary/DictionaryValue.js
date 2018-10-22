import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

const styles = theme => ({
  root: {
    display: "inline-grid",
    paddingLeft: theme.spacing.unit,
    userSelect: "text",
  },
  limit: {
    maxWidth: "60%",
  },
  explicitRightMargin: { marginRight: "24px" },
});

class DictionaryValue extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    constrain: PropTypes.bool,
    explicitRightMargin: PropTypes.bool,
  };

  static defaultProps = {
    className: null,
    constrain: false,
    explicitRightMargin: false,
  };

  render() {
    const {
      className: classNameProp,
      classes,
      children,
      constrain,
      explicitRightMargin,
      ...props
    } = this.props;
    const className = classnames(classes.root, classNameProp, {
      [classes.limit]: constrain,
      [classes.explicitRightMargin]: explicitRightMargin,
    });

    return (
      <Typography
        component="td"
        variant="body1"
        className={className}
        {...props}
      >
        {children}
      </Typography>
    );
  }
}

export default withStyles(styles)(DictionaryValue);
