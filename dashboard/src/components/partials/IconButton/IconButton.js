import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";

const styles = () => ({
  root: {
    paddingRight: 4,
  },
});

class IconButton extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
    icon: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
  };

  render() {
    const { children, icon, classes, ...props } = this.props;

    return (
      <Button {...props} className={classes.root}>
        {children} {icon}
      </Button>
    );
  }
}

export default withStyles(styles)(IconButton);
