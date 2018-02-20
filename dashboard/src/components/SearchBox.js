import React from "react";
import PropTypes from "prop-types";

import { withStyles } from "material-ui/styles";
// import Typography from "material-ui/Typography";
import filterIcon from "material-ui-icons/FilterList";

const styles = theme => ({
  box: {
    display: "flex",
    border: "1px solid",
    borderColor: theme.palette.divider,
    backgroundColor: "white",
    boxShadow:
      "0 0 2px 0 rgba(0,0,0,0.14), 0 2px 2px 0 rgba(0,0,0,0.12), 0 1px 3px 0 rgba(0,0,0,0.20)",
  },
  filterIconContainer: {
    display: "flex",
    padding: "6px 8px 8px",
    height: 36,
  },
  textField: {
    borderRadius: 3,
    width: 300,
    height: 36,
    fontSize: 14,
    border: "none",
    "&:focus": { outline: "none" },
  },
  save: {
    alignSelf: "center",
    marginRight: 8,
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
      <div className={classes.box}>
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
      </div>
    );
  }
}

export default withStyles(styles)(SearchBox);
