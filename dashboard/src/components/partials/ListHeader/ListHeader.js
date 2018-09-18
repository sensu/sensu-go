import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import Checkbox from "@material-ui/core/Checkbox";

import AppLayout from "/components/AppLayout";

const styles = theme => ({
  root: {
    // This padding is set to match the "checkbox" padding of a <TableCell>
    // component to keep the header checkbox aligned with row checkboxes.
    // See: https://github.com/mui-org/material-ui/blob/3353f44/packages/material-ui/src/TableCell/TableCell.js#L50
    paddingLeft: 12,
    paddingRight: 12,
    paddingTop: theme.spacing.unit / 2,
    paddingBottom: theme.spacing.unit / 2,
    backgroundColor: theme.palette.primary.light,
    color: theme.palette.primary.contrastText,
    display: "flex",
    alignItems: "center",
    zIndex: 1,
    // "& *": {
    //   color: theme.palette.primary.contrastText,
    // },
  },
  active: {
    backgroundColor: theme.palette.primary.main,
  },
  sticky: {
    position: "sticky",
    color: theme.palette.primary.contrastText,
  },

  grow: {
    flex: "1 1 auto",
  },
});

class ListHeader extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    sticky: PropTypes.bool,

    selectedCount: PropTypes.number.isRequired,
    rowCount: PropTypes.number.isRequired,
    onClickSelect: PropTypes.func,

    renderActions: PropTypes.func,
    renderBulkActions: PropTypes.func,
  };

  static defaultProps = {
    sticky: false,
    renderActions: () => {},
    renderBulkActions: () => {},
    onClickSelect: () => {},
  };

  render() {
    const {
      sticky,
      classes,
      onClickSelect,
      selectedCount,
      rowCount,
      renderActions,
      renderBulkActions,
    } = this.props;

    return (
      <AppLayout.Context.Consumer>
        {({ topBarHeight }) => (
          <Typography
            component="div"
            className={classnames(classes.root, {
              [classes.active]: selectedCount > 0,
              [classes.sticky]: sticky,
            })}
            style={{ top: sticky ? topBarHeight : undefined }}
          >
            <Checkbox
              component="button"
              onClick={onClickSelect}
              checked={selectedCount === rowCount}
              indeterminate={selectedCount > 0 && selectedCount !== rowCount}
              style={{ color: "inherit" }}
            />
            {selectedCount > 0 && <div>{selectedCount} Selected</div>}
            <div className={classes.grow} />
            {selectedCount > 0 ? renderBulkActions() : renderActions()}
          </Typography>
        )}
      </AppLayout.Context.Consumer>
    );
  }
}

export default withStyles(styles)(ListHeader);
