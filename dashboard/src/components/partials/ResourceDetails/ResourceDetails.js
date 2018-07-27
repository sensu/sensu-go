import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    paddingTop: 10,
    paddingBottom: 10,
    display: "flex",
    flexDirection: "row",
  },
  item: {
    lineHeight: "26px",
    color: theme.palette.text.secondary,
    "& strong": {
      color: theme.palette.text.primary,
    },
    "& a, & a:hover, & a:visited": {
      color: "inherit",
    },
  },
  title: {
    fontWeight: 600,
  },
  icon: {
    marginRight: 24,
    flex: "none",
  },
});

class ResourceDetails extends React.PureComponent {
  static propTypes = {
    icon: PropTypes.node,
    title: PropTypes.node,
    details: PropTypes.node,
    classes: PropTypes.object.isRequired,
  };

  static defaultProps = {
    icon: undefined,
    title: undefined,
    details: undefined,
  };

  render() {
    const { icon, title, details, classes } = this.props;

    return (
      <div className={classes.root}>
        {icon && <div className={classes.icon}>{icon}</div>}
        <div className={classes.item}>
          <div className={classes.title}>{title}</div>
          <div className={classes.details}>{details}</div>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(ResourceDetails);
