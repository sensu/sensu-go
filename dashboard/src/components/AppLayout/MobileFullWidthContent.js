import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
    alignItems: "stretch",

    // TODO: Replace this reference to the layout gutter width with a value
    // provided through a react context.
    // (see https://github.com/sensu/sensu-go/issues/2002)
    marginLeft: -theme.spacing.unit,
    marginRight: -theme.spacing.unit,

    [theme.breakpoints.up("md")]: {
      marginLeft: 0,
      marginRight: 0,
    },
  },
});

class MobileFullWidthContent extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    children: PropTypes.node,
  };

  static defaultProps = {
    children: undefined,
  };

  render() {
    const { classes, children } = this.props;

    return <div className={classes.root}>{children}</div>;
  }
}

export default withStyles(styles)(MobileFullWidthContent);
