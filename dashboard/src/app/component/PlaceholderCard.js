import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import Card from "@material-ui/core/Card";

const styles = () => ({
  root: {
    display: "flex",
    minHeight: 180,
    alignItems: "center",
    justifyContent: "center",
  },
  tall: {
    minHeight: 360,
  },
});

class PlaceholderCard extends React.Component {
  static propTypes = {
    tall: PropTypes.bool,
    ...Card.propTypes,
  };

  static defaultProps = {
    tall: false,
  };

  render() {
    const { classes, tall, children, ...props } = this.props;
    const className = classnames(classes.root, { [classes.tall]: tall });

    return (
      <Typography
        component={Card}
        variant="body1"
        className={className}
        {...props}
      >
        [ placeholder ]
      </Typography>
    );
  }
}

export default withStyles(styles)(PlaceholderCard);
