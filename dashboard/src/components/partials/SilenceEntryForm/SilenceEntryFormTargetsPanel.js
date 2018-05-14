import React from "react";
import { Field } from "@10xjs/form";

import Table, {
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from "material-ui/Table";

import Panel from "./SilenceEntryFormPanel";

class SilenceEntryFormTargetsPanel extends React.PureComponent {
  render() {
    return (
      <Field
        path="targets"
        format={value => (value === undefined ? [] : value)}
      >
        {targets => (
          <Panel
            title="Targets"
            summary={`${targets.props.value.length} targets selected`}
            hasDefaultValue={false}
          >
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Check</TableCell>
                  <TableCell>Subscription</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {targets.props.value.map(target => (
                  <TableRow key={`${target.subscription}:${target.check}`}>
                    <TableCell>{target.check || "*"}</TableCell>
                    <TableCell>{target.subscription || "*"}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Panel>
        )}
      </Field>
    );
  }
}

export default SilenceEntryFormTargetsPanel;
