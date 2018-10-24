import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";

import { withStyles } from "@material-ui/core/styles";
import Paper from "@material-ui/core/Paper";
import Typography from "@material-ui/core/Typography";
import Icon from "@material-ui/icons/FilterList";
import IconButton from "@material-ui/core/IconButton";
import ClearIcon from "@material-ui/icons/Cancel";
import Slide from "@material-ui/core/Slide";

const Keycap = withStyles(theme => ({
  root: {
    // monospaced fonts appear displayed larger
    fontSize: "0.71rem",
    lineHeight: "inherit",
    fontFamily: theme.typography.monospace.fontFamily,
    paddingLeft: theme.spacing.unit / 2,
    paddingRight: theme.spacing.unit / 2,
    border: `1px solid ${theme.palette.divider}`,
    borderBottomWidth: 2,
    borderRadius: theme.spacing.unit / 2,
  },
}))(({ classes, ...props }) => (
  <Typography
    component="span"
    type="caption"
    className={classes.root}
    {...props}
  />
));

const SearchInput = withStyles(theme => ({
  root: {
    width: "100%",
    backgroundColor: "inherit",
    "&:focus": {
      outline: "none",
    },
    "&::placeholder": {
      color: theme.palette.text.hint,
    },
  },
}))(({ classes, ...props }) => (
  <input type="search" className={classes.root} {...props} />
));

const styles = theme => ({
  root: {
    overflow: "hidden",
    minWidth: 300,
    display: "flex",
    border: "1px solid",
    borderColor: theme.palette.divider,
    alignItems: "center",
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
    height: theme.spacing.unit * 5,
  },
  iconContainer: {
    display: "flex",
    alignItems: "center",
    paddingRight: theme.spacing.unit,
    color: theme.palette.action.active,
  },
  inputContainer: {
    display: "flex",
    alignItems: "center",
    width: "100%",
    fontFamily: theme.typography.monospace.fontFamily,
  },
});

const defaultState = {
  value: null,
  focus: false,
};

class SearchBox extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    onSearch: PropTypes.func.isRequired,
    initialValue: PropTypes.string,
    placeholder: PropTypes.string,
  };

  static defaultProps = {
    className: "",
    placeholder: "Search…",
    initialValue: "",
  };

  state = defaultState;

  componentWillReceiveProps(nextProps) {
    if (nextProps.initialValue !== this.props.initialValue) {
      this.resetState();
    }
  }

  resetState() {
    this.setState(defaultState);
  }

  clearFilter = () => {
    this.resetState();
    this.props.onSearch("");
  };

  handleChange = ev => {
    this.setState({ value: ev.currentTarget.value || "" });
  };

  handleKeyPress = ev => {
    if (ev.key === "Enter") {
      if (this.state.value !== null) {
        this.props.onSearch(this.state.value);
        this.resetState();
      }

      ev.currentTarget.blur();
    }
  };

  handleFocus = () => {
    this.setState({ focus: true });
  };

  handleBlur = () => {
    this.setState({ focus: false });
  };

  render() {
    const {
      classes,
      className: classNameProp,
      initialValue,
      placeholder,
      onSearch,
      ...props
    } = this.props;

    const className = classnames(classNameProp, classes.root);

    let value = this.state.value;
    if (value === null) {
      value = initialValue || "";
    }

    return (
      <Paper className={className} {...props}>
        <div className={classes.iconContainer}>
          <Icon className={classes.filterIcon} />
        </div>
        <Typography
          component="div"
          variant="body1"
          className={classes.inputContainer}
        >
          <SearchInput
            placeholder={placeholder}
            onChange={this.handleChange}
            onKeyPress={this.handleKeyPress}
            onFocus={this.handleFocus}
            onBlur={this.handleBlur}
            value={value}
          />
          {this.state.focus && <Keycap>return</Keycap>}
          <Slide
            direction="left"
            in={value.length > 0}
            timeout={200}
            mountOnEnter
            unmountOnExit
          >
            <IconButton onClick={this.clearFilter}>
              <ClearIcon />
            </IconButton>
          </Slide>
        </Typography>
      </Paper>
    );
  }
}

export default withStyles(styles)(SearchBox);
