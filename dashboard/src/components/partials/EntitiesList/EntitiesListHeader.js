import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { withStyles } from "@material-ui/core/styles";

import Button from "@material-ui/core/Button";
import Checkbox from "@material-ui/core/Checkbox";
import ListItemText from "@material-ui/core/ListItemText";
import MenuItem from "@material-ui/core/MenuItem";
import Typography from "@material-ui/core/Typography";

import { TableListHeader, TableListSelect } from "/components/TableList";
import ButtonSet from "/components/ButtonSet";

import ConfirmDelete from "/components/partials/ConfirmDelete";

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
    onClickSelect: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    selectedCount: PropTypes.number.isRequired,
    onChangeQuery: PropTypes.func.isRequired,
    environment: PropTypes.object,
  };

  static defaultProps = {
    onClickSelect: () => {},
    onChangeFilter: () => {},
    onChangeSort: () => {},
    onSubmitDelete: () => {},
    selectedCount: 0,
    environment: undefined,
  };

  static fragments = {
    environment: gql`
      fragment EntitiesListHeader_environment on Environment {
        subscriptions(orderBy: OCCURRENCES, omitEntity: true) {
          values(limit: 25)
        }
      }
    `,
  };

  _handleChangeSort = val => {
    let newVal = val;
    this.props.onChangeQuery(query => {
      // Toggle between ASC & DESC
      const curVal = query.get("order");
      if (curVal === "ID" && newVal === "ID") {
        newVal = "ID_DESC";
      }
      query.set("order", newVal);
    });
  };

  _handleChangeFiler = (filter, val) => {
    switch (filter) {
      case "subscription":
        this.props.onChangeQuery({ filter: `'${val}' IN Subscriptions` });
        break;
      default:
        throw new Error(`unexpected filter '${filter}'`);
    }
  };

  render() {
    const {
      classes,
      environment,
      onClickDelete,
      onClickSelect,
      selectedCount,
    } = this.props;

    const subscriptions = environment ? environment.subscriptions.values : [];

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
        {selectedCount > 0 ? (
          <ButtonSet>
            <ConfirmDelete
              identifier={`${selectedCount} ${
                selectedCount === 1 ? "entity" : "entities"
              }`}
              onSubmit={onClickDelete}
            >
              {confirm => (
                <Button onClick={confirm.open}>
                  <Typography variant="button">Delete</Typography>
                </Button>
              )}
            </ConfirmDelete>
          </ButtonSet>
        ) : (
          <ButtonSet>
            <TableListSelect
              label="subscription"
              // eslint-disable-next-line react/jsx-no-bind
              onChange={this._handleChangeFiler.bind(this, "subscription")}
            >
              {subscriptions.map(entry => (
                <MenuItem key={entry} value={entry}>
                  <ListItemText primary={entry} />
                </MenuItem>
              ))}
            </TableListSelect>
            <TableListSelect label="Sort" onChange={this._handleChangeSort}>
              <MenuItem key="ID" value="ID">
                <ListItemText>Name</ListItemText>
              </MenuItem>
              <MenuItem key="LASTSEEN" value="LASTSEEN">
                <ListItemText>Last Seen</ListItemText>
              </MenuItem>
            </TableListSelect>
          </ButtonSet>
        )}
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
