from pathlib import Path
import argparse
import sys

from ml_service.pipeline.pipeline import run_pipeline


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run DataForge ML analysis pipeline for image classification dataset."
    )

    parser.add_argument(
        "--dataset",
        required=True,
        help="Path to dataset.csv",
    )

    parser.add_argument(
        "--images-dir",
        required=True,
        help="Path to directory with images",
    )

    parser.add_argument(
        "--output-dir",
        required=True,
        help="Path to output directory",
    )

    return parser.parse_args()


def main() -> None:
    args = parse_args()

    dataset_path = Path(args.dataset)
    images_dir = Path(args.images_dir)
    output_dir = Path(args.output_dir)

    if not dataset_path.exists():
        print(f"ERROR: dataset file not found: {dataset_path}")
        sys.exit(1)

    if not images_dir.exists():
        print(f"ERROR: images directory not found: {images_dir}")
        sys.exit(1)

    result = run_pipeline(
        dataset_path=dataset_path,
        images_dir=images_dir,
        output_dir=output_dir,
    )

    print("\nCLI run completed.")
    print(f"Status: {result['status']}")
    print(f"Objects: {result['objects_count']}")
    print(f"Dataset v2 objects: {result['dataset_v2_objects_count']}")
    print(f"Review queue objects: {result['review_queue_count']}")
    print(f"Dataset readiness: {result['dataset_readiness_score']:.2f}/100")
    print(f"Output dir: {result['output_dir']}")

    print("\nGenerated files:")
    for name, path in result["files"].items():
        print(f"- {name}: {path}")


if __name__ == "__main__":
    main()