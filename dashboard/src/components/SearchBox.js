import React from "react";
import PropTypes from "prop-types";

import { withStyles } from "material-ui/styles";
import Paper from "material-ui/Paper";
// import Typography from "material-ui/Typography";
import filterIcon from "material-ui-icons/FilterList";

const styles = theme => ({
  box: {
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
    width: 300,
    height: 36,
    fontSize: 14,
    border: "none",
    backgroundColor: theme.palette.background.paper,
    "&:focus": { outline: "none" },
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
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    FilterIcon: PropTypes.func.isRequired,
    onUpdateInput: PropTypes.func.isRequired,
    state: PropTypes.string.isRequired,
  };

  static defaultProps = {
    FilterIcon: filterIcon,
  };

  updateInput = event => {
    this.props.onUpdateInput(event.currentTarget.value);
  };

  render() {
    const { classes, FilterIcon } = this.props;

    return (
      <Paper className={classes.box}>
        <div className={classes.filterIconContainer}>
          <FilterIcon className={classes.filterIcon} />
        </div>
        <input
          id="search"
          type="text"
          placeholder={"Filter all events"}
          className={classes.textField}
          value={this.props.state}
          onChange={this.updateInput}
        />
        {/* <Typography className={classes.save} type="button">
          Save
        </Typography> */}
      </Paper>
    );
  }
}

export default withStyles(styles)(SearchBox);
