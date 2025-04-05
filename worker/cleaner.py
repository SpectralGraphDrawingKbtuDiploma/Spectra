import sys

def clean_matrix_market(input_filename, output_filename):
    with open(input_filename, 'r') as fin:
        lines = fin.readlines()
    non_comment_lines = [line.strip() for line in lines if line.strip() and not line.strip().startswith('%')]

    if not non_comment_lines:
        print("No non-comment lines found in the input file.")
        return

    output_lines = []

    # dims = non_comment_lines[0].split()
    # if len(dims) < 2:
    #     print("Error: the first non-comment line does not have enough entries.")
    #     return
    # output_lines.append(" ".join(dims[:2]))

    for line in non_comment_lines[1:]:
        parts = line.split()
        if len(parts) >= 2:
            output_lines.append(" ".join(parts[:2]))

    with open(output_filename, 'w') as fout:
        for line in output_lines:
            fout.write(line + "\n")

    print(f"Cleaned file with only first two columns has been saved to '{output_filename}'.")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python clean.py <input_matrix_market_file.mtx> <output_file.txt>")
        sys.exit(1)

    input_filename = sys.argv[1]
    output_filename = sys.argv[2]
    clean_matrix_market(input_filename, output_filename)