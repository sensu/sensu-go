import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Button from "@material-ui/core/Button";
import capitalize from "lodash/capitalize";

import Typography from "@material-ui/core/Typography";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";

import CollapsingMenu from "/components/partials/CollapsingMenu";
import ButtonSet from "/components/ButtonSet";
import Menu from "@material-ui/core/Menu";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import StatusMenu from "/components/partials/StatusMenu";
import ListHeader from "/components/partials/ListHeader";

class EventsListHeader extends React.PureComponent {
  static propTypes = {
    onClickClearSilences: PropTypes.func.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    onClickResolve: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    selectedItems: PropTypes.array.isRequired,
    rowCount: PropTypes.number.isRequired,
    environment: PropTypes.shape({
      checks: PropTypes.object,
      entities: PropTypes.object,
    }),
    onChangeQuery: PropTypes.func.isRequired,
  };

  static defaultProps = {
    environment: null,
  };

  static fragments = {
    event: gql`
      fragment EventsListHeader_event on Event {
        check {
          isSilenced
        }
      }
    `,
    environment: gql`
      fragment EventsListHeader_environment on Environment {
        checks(limit: 1000) {
          nodes {
            name
          }
        }

        entities(limit: 1000) {
          nodes {
            name
          }
        }
      }
    `,
  };

  requeryEntity = newValue => {
    this.props.onChangeQuery({ filter: `Entity.ID == '${newValue}'` });
  };

  requeryCheck = newValue => {
    this.props.onChangeQuery({ filter: `Check.Name == '${newValue}'` });
  };

  requeryHide = newValue => {
    if (newValue === "passing") {
      this.props.onChangeQuery({ filter: `Check.Status != 0` });
    } else if (newValue === "silenced") {
      this.props.onChangeQuery({ filter: `!IsSilenced` });
    } else {
      throw new TypeError(`unknown value ${newValue}`);
    }
  };

  requeryStatus = newValue => {
    if (Array.isArray(newValue)) {
      if (newValue.length === 1) {
        this.props.onChangeQuery({ filter: `Check.Status == ${newValue}` });
      } else {
        const val = newValue.join(",");
        this.props.onChangeQuery({ filter: `Check.Status IN (${val})` });
      }
    } else {
      this.props.onChangeQuery(query => query.delete("filter"));
    }
  };

  _handleChangeSort = newValue => {
    this.props.onChangeQuery({ order: newValue });
  };

  render() {
    const {
      selectedItems,
      rowCount,
      onClickClearSilences,
      onClickDelete,
      onClickSelect,
      onClickSilence,
      onClickResolve,
      environment,
    } = this.props;

    const entityNames = environment
      ? environment.entities.nodes.map(node => node.name)
      : [];

    const checkNames = [
      ...(environment ? environment.checks.nodes.map(node => node.name) : []),
      "keepalive",
    ];

    const selectedCount = selectedItems.length;
    const selectedSilenced = selectedItems.filter(ev => ev.check.isSilenced);

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
                selectedCount === 1 ? "event" : "events"
              }`}
              onSubmit={onClickDelete}
            >
              {confirm => (
                <Button onClick={confirm.open}>
                  <Typography variant="button">Delete</Typography>
                </Button>
              )}
            </ConfirmDelete>
            <Button onClick={onClickSilence}>
              <Typography variant="button">Silence</Typography>
            </Button>
            {selectedSilenced.length > 0 && (
              <Button onClick={onClickClearSilences}>
                <Typography variant="button">Unsilence</Typography>
              </Button>
            )}
            <Button onClick={onClickResolve}>
              <Typography variant="button">Resolve</Typography>
            </Button>
          </ButtonSet>
        )}
        renderActions={() => (
          <CollapsingMenu breakpoint="md">
            <CollapsingMenu.SubMenu
              title="Hide"
              renderMenu={({ anchorEl, close }) => (
                <Menu open onClose={close} anchorEl={anchorEl}>
                  <MenuItem
                    onClick={() => {
                      this.requeryHide("passing");
                      close();
                    }}
                  >
                    <ListItemText primary="Passing" />
                  </MenuItem>
                  <MenuItem
                    onClick={() => {
                      this.requeryHide("silenced");
                      close();
                    }}
                  >
                    <ListItemText primary="Silenced" />
                  </MenuItem>
                </Menu>
              )}
            />
            <CollapsingMenu.SubMenu
              title="Entity"
              renderMenu={({ anchorEl, close }) => (
                <Menu open onClose={close} anchorEl={anchorEl}>
                  {entityNames.map(name => (
                    <MenuItem
                      key={name}
                      onClick={() => {
                        this.requeryEntity(name);
                        close();
                      }}
                    >
                      <ListItemText primary={name} />
                    </MenuItem>
                  ))}
                </Menu>
              )}
            />
            <CollapsingMenu.SubMenu
              title="Check"
              renderMenu={({ anchorEl, close }) => (
                <Menu open onClose={close} anchorEl={anchorEl}>
                  {checkNames.map(name => (
                    <MenuItem
                      key={name}
                      onClick={() => {
                        this.requeryCheck(name);
                        close();
                      }}
                    >
                      <ListItemText primary={name} />
                    </MenuItem>
                  ))}
                </Menu>
              )}
            />
            <CollapsingMenu.SubMenu
              title="Status"
              pinned
              renderMenu={({ anchorEl, close }) => (
                <StatusMenu
                  anchorEl={anchorEl}
                  onClose={close}
                  onChange={val => {
                    this.requeryStatus(val);
                    close();
                  }}
                />
              )}
            />
            <CollapsingMenu.SubMenu
              title="Sort"
              pinned
              renderMenu={({ anchorEl, close }) => (
                <Menu open onClose={close} anchorEl={anchorEl}>
                  {["SEVERITY", "NEWEST", "OLDEST"].map(name => (
                    <MenuItem
                      key={name}
                      onClick={() => {
                        this._handleChangeSort(name);
                        close();
                      }}
                    >
                      <ListItemText primary={capitalize(name)} />
                    </MenuItem>
                  ))}
                </Menu>
              )}
            />
          </CollapsingMenu>
        )}
      />
    );
  }
}

export default EventsListHeader;
