import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import Checkbox from "@material-ui/core/Checkbox";

const styles = theme => {
  const toolbar = theme.mixins.toolbar;
  const xsBrk = `${theme.breakpoints.up("xs")} and (orientation: landscape)`;
  const smBrk = theme.breakpoints.up("sm");
  const calcTopWithFallback = size => ({
    top: `calc(${size}px + env(safe-area-inset-top))`,
    fallbacks: [{ top: size }],
  });

  return {
    root: {
      // This padding is set to match the "checkbox" padding of a <TableCell>
      // component to keep the header checkbox aligned with row checkboxes.
      // See: https://github.com/mui-org/material-ui/blob/3353f44/packages/material-ui/src/TableCell/TableCell.js#L50
      paddingLeft: 12,
      paddingRight: 24,
      paddingTop: 4,
      paddingBottom: 4,
      backgroundColor: theme.palette.primary.light,
      color: theme.palette.primary.contrastText,
      display: "flex",
      alignItems: "center",
      zIndex: theme.zIndex.appBar - 1,
      "& *": {
        color: theme.palette.primary.contrastText,
      },
    },
    active: {
      backgroundColor: theme.palette.primary.main,
    },
    sticky: {
      position: "sticky",
      ...calcTopWithFallback(toolbar.minHeight),
      [xsBrk]: {
        ...calcTopWithFallback(toolbar[xsBrk].minHeight),
      },
      [smBrk]: {
        ...calcTopWithFallback(toolbar[smBrk].minHeight),
      },
      color: theme.palette.primary.contrastText,
    },

    grow: {
      flex: "1 1 auto",
    },
  };
};

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
      <Typography
        component="div"
        className={classnames(classes.root, {
          [classes.active]: selectedCount > 0,
          [classes.sticky]: sticky,
        })}
      >
        <Checkbox
          component="button"
          onClick={onClickSelect}
          checked={selectedCount === rowCount}
          indeterminate={selectedCount > 0 && selectedCount !== rowCount}
        />
        {selectedCount > 0 && <div>{selectedCount} Selected</div>}
        <div className={classes.grow} />
        {selectedCount > 0 ? renderBulkActions() : renderActions()}
      </Typography>
    );
  }
}

export default withStyles(styles)(ListHeader);
