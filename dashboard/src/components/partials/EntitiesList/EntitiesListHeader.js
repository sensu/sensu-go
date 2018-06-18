import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import ButtonSet from "/components/ButtonSet";
import Checkbox from "@material-ui/core/Checkbox";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import ListItemText from "@material-ui/core/ListItemText";
import MenuItem from "@material-ui/core/MenuItem";
import { TableListHeader, TableListSelect } from "/components/TableList";

const styles = theme => ({
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
    classes: PropTypes.object.isRequired,
    onClickSelect: PropTypes.func,
    onChangeFilter: PropTypes.func,
    onChangeSort: PropTypes.func,
    onSubmitDelete: PropTypes.func,
    selectedCount: PropTypes.number,
    subscriptions: PropTypes.shape({ values: PropTypes.array }),
  };

  static defaultProps = {
    onClickSelect: () => {},
    onChangeFilter: () => {},
    onChangeSort: () => {},
    onSubmitDelete: () => {},
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

  _renderListActions = () => {
    const {
      onChangeFilter,
      onChangeSort,
      subscriptions: subscriptionsProp,
    } = this.props;
    const subscriptions = subscriptionsProp ? subscriptionsProp.values : [];

    return (
      <ButtonSet>
        <TableListSelect
          label="subscription"
          onChange={val => onChangeFilter("subscription", val)}
        >
          {subscriptions.map(entry => (
            <MenuItem key={entry} value={entry}>
              <ListItemText primary={entry} />
            </MenuItem>
          ))}
        </TableListSelect>
        <TableListSelect label="Sort" onChange={onChangeSort}>
          <MenuItem key="ID" value="ID">
            <ListItemText>Name</ListItemText>
          </MenuItem>
          <MenuItem key="LASTSEEN" value="LASTSEEN">
            <ListItemText>Last Seen</ListItemText>
          </MenuItem>
        </TableListSelect>
      </ButtonSet>
    );
  };

  _renderBulkActions = () => {
    const { onSubmitDelete, selectedCount } = this.props;
    const resource = selectedCount === 1 ? "entity" : "entities";
    const ref = `${selectedCount} ${resource}`;

    return (
      <ButtonSet>
        <ConfirmDelete identifier={ref} onSubmit={onSubmitDelete}>
          {confirm => <Button onClick={confirm.open}>Delete</Button>}
        </ConfirmDelete>
      </ButtonSet>
    );
  };

  render() {
    const { classes, onClickSelect, selectedCount } = this.props;

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
        {selectedCount === 0
          ? this._renderListActions()
          : this._renderBulkActions()}
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
