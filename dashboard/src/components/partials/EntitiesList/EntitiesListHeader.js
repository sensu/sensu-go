import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Button from "@material-ui/core/Button";
import ListItemText from "@material-ui/core/ListItemText";
import MenuItem from "@material-ui/core/MenuItem";
import Typography from "@material-ui/core/Typography";

import ButtonSet from "/components/ButtonSet";
import ListHeader from "/components/partials/ListHeader";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import ButtonMenu from "/components/partials/ButtonMenu";

class EntitiesListHeader extends React.PureComponent {
  static propTypes = {
    onClickSelect: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    selectedCount: PropTypes.number.isRequired,
    rowCount: PropTypes.number.isRequired,
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
      environment,
      onClickDelete,
      onClickSelect,
      selectedCount,
      rowCount,
    } = this.props;

    const subscriptions = environment ? environment.subscriptions.values : [];

    return (
      <ListHeader
        sticky
        selectedCount={selectedCount}
        rowCount={rowCount}
        onClickSelect={onClickSelect}
        renderBulkActions={() => (
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
        )}
        renderActions={() => (
          <ButtonSet>
            <ButtonMenu
              label="subscription"
              // eslint-disable-next-line react/jsx-no-bind
              onChange={this._handleChangeFiler.bind(this, "subscription")}
            >
              {subscriptions.map(entry => (
                <MenuItem key={entry} value={entry}>
                  <ListItemText primary={entry} />
                </MenuItem>
              ))}
            </ButtonMenu>
            <ButtonMenu label="Sort" onChange={this._handleChangeSort}>
              <MenuItem key="ID" value="ID">
                <ListItemText>Name</ListItemText>
              </MenuItem>
              <MenuItem key="LASTSEEN" value="LASTSEEN">
                <ListItemText>Last Seen</ListItemText>
              </MenuItem>
            </ButtonMenu>
          </ButtonSet>
        )}
      />
    );
  }
}

export default EntitiesListHeader;
