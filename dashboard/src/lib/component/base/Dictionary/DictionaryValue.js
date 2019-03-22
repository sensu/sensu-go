import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import DictionaryEntry from "./DictionaryEntry";

const styles = theme => ({
  root: {
    display: "table-cell",
    userSelect: "text",
  },
  limit: {
    maxWidth: "60%",
  },
  padding: {
    paddingLeft: theme.spacing.unit,
  },
  scrollableContent: { display: "inline-grid" },
  explicitRightMargin: { paddingRight: "24px" },
});

class DictionaryValue extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    constrain: PropTypes.bool,
    explicitRightMargin: PropTypes.bool,
    scrollableContent: PropTypes.bool,
  };

  static defaultProps = {
    className: null,
    constrain: false,
    explicitRightMargin: false,
    scrollableContent: false,
  };

  renderCell = ({ fullWidth }) => {
    const {
      className: classNameProp,
      classes,
      children,
      constrain,
      explicitRightMargin,
      scrollableContent,
      ...props
    } = this.props;
    const className = classnames(classes.root, classNameProp, {
      [classes.limit]: constrain,
      [classes.explicitRightMargin]: explicitRightMargin,
      [classes.scrollableContent]: scrollableContent,
      [classes.padding]: !fullWidth,
    });

    return (
      <Typography
        component="td"
        variant="body2"
        className={className}
        colSpan={fullWidth ? 2 : 1}
        {...props}
      >
        {children}
      </Typography>
    );
  };

  render() {
    return (
      <DictionaryEntry.Context>
        {contextProps => this.renderCell(contextProps)}
      </DictionaryEntry.Context>
    );
  }
}

export default withStyles(styles)(DictionaryValue);
