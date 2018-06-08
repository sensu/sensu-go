import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import { withStyles } from "@material-ui/core/styles";
import { TableListHeader, TableListSelect } from "/components/TableList";
import ButtonSet from "/components/ButtonSet";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";

const styles = theme => ({
  filterActions: {
    display: "none",
    [theme.breakpoints.up("sm")]: {
      display: "flex",
    },
  },
  // Remove padding from button container
  checkbox: {
    marginLeft: -11,
    color: theme.palette.primary.contrastText,
  },
  grow: {
    flex: "1 1 auto",
  },
});

class EntitiesListHeader extends React.PureComponent {
  static propTypes = {
    actions: PropTypes.node.isRequired,
    bulkActions: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    onClickSelect: PropTypes.func,
    onChangeFilter: PropTypes.func,
    selectedCount: PropTypes.number,
    subscriptions: PropTypes.shape({ values: PropTypes.array }),
  };

  static defaultProps = {
    onClickSelect: () => {},
    onChangeFilter: () => {},
    selectedCount: 0,
    subscriptions: undefined,
  };

  static fragments = {
    subscriptions: gql`
      fragment EntitiesListHeader_subscriptions on SubscriptionSet {
        values(limit: 25)
      }
    `,
  };

  getSubscriptions = () =>
    this.props.subscriptions ? this.props.subscriptions.values : [];

  render() {
    const {
      actions,
      bulkActions,
      classes,
      onClickSelect,
      onChangeFilter,
      selectedCount,
    } = this.props;

    return (
      <TableListHeader sticky active={selectedCount > 0}>
        <Checkbox
          component="button"
          className={classes.checkbox}
          onClick={onClickSelect}
          checked={false}
          indeterminate={selectedCount > 0}
        />
        {selectedCount > 0 && <div>{selectedCount} Selected</div>}
        <div className={classes.grow} />
        <ButtonSet>
          <TableListSelect
            label="subscription"
            onChange={val => onChangeFilter("subscription", val)}
          >
            {this.getSubscriptions().map(entry => (
              <MenuItem key={entry} value={entry}>
                <ListItemText primary={entry} />
              </MenuItem>
            ))}
          </TableListSelect>
        </ButtonSet>
        {selectedCount > 0 ? bulkActions : actions}
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
