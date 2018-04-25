import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

const styles = theme => ({
  root: {
    // ...
  },
  inset: {
    paddingLeft: theme.spacing.unit * 2.5,
  },
});

class DetailedListItemSubtitle extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    inset: PropTypes.bool,
  };

  static defaultProps = {
    inset: false,
  };

  render() {
    const { classes, children, inset, ...props } = this.props;
    const className = classnames(classnames.root, { [classes.inset]: inset });

    return (
      <Typography
        variant="body1"
        color="textSecondary"
        className={className}
        {...props}
      >
        {children}
      </Typography>
    );
  }
}

export default withStyles(styles)(DetailedListItemSubtitle);
