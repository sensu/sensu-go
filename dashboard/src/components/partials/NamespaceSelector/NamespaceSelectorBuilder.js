import React from "react";
import PropTypes from "prop-types";
import classNames from "classnames";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import ArrowIcon from "@material-ui/icons/ArrowDropDown";

const styles = theme => ({
  label: {
    color: theme.palette.primary.contrastText,
    opacity: 0.9,
    display: "block",
  },
  prefixLabel: {
    fontWeight: "lighter",
    fontSize: "0.75rem",
  },
  nameLabel: {
    fontSize: "1.25rem",
  },
  nameContainer: {
    margin: "-6px 0 0",
    display: "flex",
    justifyContent: "space-between",
  },
  arrow: {
    color: theme.palette.primary.contrastText,
  },
});

class NamespaceSelectorBuilder extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    namespace: PropTypes.object,
  };

  static defaultProps = {
    namespace: null,
  };

  render() {
    const { classes, namespace } = this.props;

    let [prefix, ...name] = namespace ? namespace.name.split("-") : [];
    if (name.length === 0) {
      name = prefix;
      prefix = null;
    } else {
      name = name.join("-");
    }

    return (
      <div className={classes.selectorContainer}>
        <Typography className={classNames(classes.label, classes.prefixLabel)}>
          {prefix || " "}
        </Typography>
        <div className={classes.nameContainer}>
          <Typography className={classNames(classes.label, classes.nameLabel)}>
            {name}
          </Typography>
          <span className={classes.arrow}>
            <ArrowIcon />
          </span>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(NamespaceSelectorBuilder);
