import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

const styles = theme => ({
  root: {
    width: "40%",
    display: "table-cell",
    paddingRight: theme.spacing.unit,
    textOverflow: "ellipsis",
  },
  container: {
    minHeight: theme.spacing.unit * 3,
  },
});

class DictionaryKey extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
  };

  static defaultProps = {
    className: null,
  };

  render() {
    const { className: classNameProp, classes, children } = this.props;
    const className = classnames(classes.root, classNameProp);

    return (
      <Typography component="td" color="textSecondary" className={className}>
        <div className={classes.container}>{children}</div>
      </Typography>
    );
  }
}

export default withStyles(styles)(DictionaryKey);
