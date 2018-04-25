import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  root: {
    paddingTop: theme.spacing.unit / 2,
    "&:first-child": {
      paddingTop: 0,
    },
  },
});

class DetailedListItem extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    children: PropTypes.node.isRequired,
  };

  render() {
    const { classes, children } = this.props;
    return <li className={classes.root}>{children}</li>;
  }
}

export default withStyles(styles)(DetailedListItem);
