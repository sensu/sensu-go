import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";

import Icon from "/app/component/partial/NamespaceIcon/Icon";
import Typography from "@material-ui/core/Typography";

class NamespaceLabelBase extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    name: PropTypes.string.isRequired,
    icon: PropTypes.string.isRequired,
    colour: PropTypes.string.isRequired,
  };

  static styles = theme => ({
    container: {
      display: "flex",
      alignSelf: "center",
    },
    label: {
      color: "inherit",
      marginRight: theme.spacing.unit,
      opacity: 0.88,
    },
    heavier: {
      fontWeight: 400,
    },
    lighter: {
      fontWeight: 300,
      opacity: 0.71,
    },
  });

  render() {
    const { classes, name, icon, colour } = this.props;
    const nsComponents = name.split("-");

    return (
      <div className={classes.container}>
        <Typography className={classes.label} variant="subtitle1">
          {nsComponents.length > 1 && (
            <React.Fragment>
              <span className={classes.lighter}>{nsComponents.shift()}</span>
              {" - "}
            </React.Fragment>
          )}
          <span className={classes.heavier}>{nsComponents.join("-")}</span>
        </Typography>

        <Icon icon={icon} colour={colour} />
      </div>
    );
  }
}

const enhancer = withStyles(NamespaceLabelBase.styles);
export default enhancer(NamespaceLabelBase);
