import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import CollapsingMenu from "/components/partials/CollapsingMenu";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import DeleteIcon from "@material-ui/icons/Delete";
import ListHeader from "/components/partials/ListHeader";
import ListItemText from "@material-ui/core/ListItemText";
import ListSortMenu from "/components/partials/ListSortMenu";
import { Menu } from "/components/partials/ButtonMenu";
import MenuItem from "@material-ui/core/MenuItem";
import SilenceIcon from "/icons/Silence";
import UnsilenceIcon from "/icons/Unsilence";

class EntitiesListHeader extends React.PureComponent {
  static propTypes = {
    environment: PropTypes.object,
    onChangeQuery: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    rowCount: PropTypes.number.isRequired,
    selectedItems: PropTypes.array.isRequired,
  };

  static defaultProps = {
    onClickSelect: () => {},
    onChangeFilter: () => {},
    onChangeSort: () => {},
    onSubmitDelete: () => {},
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
      onClickClearSilences,
      onClickDelete,
      onClickSelect,
      onClickSilence,
      selectedItems,
      rowCount,
    } = this.props;

    const selectedCount = selectedItems.length;
    const selectedSilenced = selectedItems.filter(en => !en.silences.length);
    const subscriptions = environment ? environment.subscriptions.values : [];

    return (
      <ListHeader
        sticky
        selectedCount={selectedCount}
        rowCount={rowCount}
        onClickSelect={onClickSelect}
        renderBulkActions={() => (
          <CollapsingMenu>
            <CollapsingMenu.Button
              alt="Create a silence targeting selected entities."
              disabled={selectedSilenced.length === 0}
              icon={<SilenceIcon />}
              onClick={onClickSilence}
              title="Silence"
              pinned
            />
            <CollapsingMenu.Button
              title="Unsilence"
              icon={<UnsilenceIcon />}
              onClick={onClickClearSilences}
              alt="Clear silences associated with selected entities."
              disabled={selectedSilenced.length > 0}
            />
            <ConfirmDelete
              identifier={`${selectedCount} ${
                selectedCount === 1 ? "entity" : "entities"
              }`}
              onSubmit={onClickDelete}
            >
              {confirm => (
                <CollapsingMenu.Button
                  title="Delete"
                  icon={<DeleteIcon />}
                  onClick={confirm.open}
                />
              )}
            </ConfirmDelete>
          </CollapsingMenu>
        )}
        renderActions={() => (
          <CollapsingMenu>
            <CollapsingMenu.SubMenu
              title="Subscription"
              renderMenu={({ anchorEl, close }) => (
                <Menu
                  anchorEl={anchorEl}
                  onChange={val => this._handleChangeFiler("subscription", val)}
                  onClose={close}
                >
                  {subscriptions.map(entry => (
                    <MenuItem key={entry} value={entry}>
                      <ListItemText primary={entry} />
                    </MenuItem>
                  ))}
                </Menu>
              )}
            />
            <CollapsingMenu.SubMenu
              title="Sort"
              pinned
              renderMenu={({ anchorEl, close }) => (
                <ListSortMenu
                  anchorEl={anchorEl}
                  onClose={close}
                  options={["ID"]}
                  onChangeQuery={this.props.onChangeQuery}
                />
              )}
            />
          </CollapsingMenu>
        )}
      />
    );
  }
}

export default EntitiesListHeader;
