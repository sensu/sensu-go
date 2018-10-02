import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

const styles = theme => ({
  root: {
    display: "grid",
    paddingLeft: theme.spacing.unit,
    userSelect: "text",
    "& > code": {
      whiteSpace: "pre-wrap",
      marginRight: "24px",
    },
  },
  limit: {
    maxWidth: "60%",
  },
});

class DictionaryValue extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    constrain: PropTypes.bool,
  };

  static defaultProps = {
    className: null,
    constrain: false,
  };

  render() {
    const {
      className: classNameProp,
      classes,
      children,
      constrain,
      ...props
    } = this.props;
    const className = classnames(classes.root, classNameProp, {
      [classes.limit]: constrain,
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
