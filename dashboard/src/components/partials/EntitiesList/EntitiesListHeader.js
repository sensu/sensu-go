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
  headerButton: {
    marginLeft: theme.spacing.unit / 2,
    "&:first-child": {
      marginLeft: theme.spacing.unit,
    },
  },
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
    classes: PropTypes.object.isRequired,
    onClickSelect: PropTypes.func,
    onChangeFilter: PropTypes.func,
    selectedCount: PropTypes.number,
    subscriptions: PropTypes.shape({ entries: PropTypes.array }),
  };

  static defaultProps = {
    onClickSelect: () => {},
    onChangeFilter: () => {},
    selectedCount: 0,
    subscriptions: { entries: [] },
  };

  static fragments = {
    subscriptions: gql`
      fragment EntitiesListHeader_subscriptions on SubscriptionSet {
        entries(orderBy: FREQUENCY)
      }
    `,
  };

  render() {
    const {
      classes,
      subscriptions,
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
            {subscriptions &&
              subscriptions.entries.map(entry => (
                <MenuItem key={entry} value={entry}>
                  <ListItemText primary={entry} />
                </MenuItem>
              ))}
          </TableListSelect>
        </ButtonSet>
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
