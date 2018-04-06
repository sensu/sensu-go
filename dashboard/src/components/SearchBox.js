import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";

import { withStyles } from "material-ui/styles";
import Paper from "material-ui/Paper";
import Typography from "material-ui/Typography";
import Icon from "material-ui-icons/FilterList";

const styles = theme => ({
  root: {
    minWidth: 300,
    display: "flex",
    border: "1px solid",
    borderColor: theme.palette.divider,
  },
  filterIconContainer: {
    display: "flex",
    padding: "6px 8px 8px",
    height: 36,
    color: theme.palette.action.active,
  },
  textField: {
    borderRadius: 3,
    height: 36,
    fontSize: 14,
    border: "none",
    width: "100%",
    backgroundColor: theme.palette.background.paper,
    "&:focus": { outline: "none" },
    color: theme.palette.text.primary,
  },
  save: {
    alignSelf: "center",
    marginRight: theme.spacing.unit,
    color: theme.palette.primary.light,
    fontWeight: "bold",
  },
});

class SearchBox extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    onChange: PropTypes.func.isRequired,
    value: PropTypes.string.isRequired,
  };

  static defaultProps = {
    className: "",
  };

  updateInput = event => {
    this.props.onChange(event.currentTarget.value);
  };

  render() {
    const { classes, className: classNameProp, value } = this.props;
    const className = classnames(classNameProp, classes.root);

    return (
      <Paper className={className}>
        <div className={classes.filterIconContainer}>
          <Icon className={classes.filterIcon} />
        </div>
        <Typography
          component="input"
          type="text"
          variant="body1"
          placeholder={"Filter all events"}
          className={classes.textField}
          value={value}
          onChange={this.updateInput}
        />
        {/* <Typography className={classes.save} variant="button">
          Save
        </Typography> */}
      </Paper>
    );
  }
}

export default withStyles(styles)(SearchBox);
